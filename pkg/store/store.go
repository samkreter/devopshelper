package store

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/samkreter/go-core/log"
	"github.com/pkg/errors"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/samkreter/devopshelper/pkg/types"
)

const (
	defaultRepositoryCollectionName = "repositories"
	defaultReviewerCollectionName = "reviewers"
)

var (
	// ErrNotFound the error is not found
	ErrNotFound = errors.New("record not found")

	ErrTransactionAborted = errors.New("transaction aborted")
)


// RepositoryStore holds information for a repository
type RepositoryStore interface {
	PopLRUReviewer(ctx context.Context, alias []string) (*types.Reviewer, error)
	GetLRUReviewer(ctx context.Context, alias []string) (*types.Reviewer, error)
	AddReviewer(ctx context.Context, reviewer *types.Reviewer) error
	GetReviewer(ctx context.Context, alias string) (*types.Reviewer, error)
	GetReviewerByADOID(ctx context.Context, adoID string) (*types.Reviewer, error)
	UpdateReviewer(ctx context.Context, reviewer *types.Reviewer) error

	// Repository Ops
	AddRepository(ctx context.Context, repo *types.Repository) error
	UpdateRepository(ctx context.Context, id string, repository *types.Repository) error
	DeleteRepository(ctx context.Context, id string) error
	GetRepositoryByID(ctx context.Context, id string) (*types.Repository, error)
	GetAllRepositories(ctx context.Context) ([]*types.Repository, error)
	GetRepositoryByName(ctx context.Context, name, project string) (*types.Repository, error)
}

// Validate the interface implementation
var _ RepositoryStore = &MongoStore{}

// MongoStoreOptions options for a mongo store
type MongoStoreOptions struct {
	UseSSL               bool
	MongoURI             string
	DBName               string
	RepositoryCollection string
	BaseGroupCollection  string
	ReviewerCollection   string
}

// MongoStore implementation to interact with a mongo database
type MongoStore struct {
	ReviewerGroupCollectionName string
	sesson                      *mgo.Session
	Options                     *MongoStoreOptions
	reviewerWriteLock sync.Mutex
}

// Close closes a mongo store and it's session
func (ms *MongoStore) Close() {
	ms.sesson.Close()
}

// NewMongoStore creates a new mongo store
func NewMongoStore(o *MongoStoreOptions) (*MongoStore, error) {
	ctx := context.Background()
	logger := log.G(ctx)
	if o.DBName == "" {
		return nil, errors.New("missing Mongo DBName")
	}

	if o.MongoURI == "" {
		return nil, errors.New("missing Mongo connection string")
	}

	if o.ReviewerCollection == "" {
		o.ReviewerCollection = defaultReviewerCollectionName
	}

	if o.RepositoryCollection == "" {
		o.RepositoryCollection = defaultRepositoryCollectionName
	}

	logger.Infof("MongoStore: Using DB: '%s' for mongo with RepoCollection: %s, BaseGroupCollection: %s, ReviewerCollection: %s",
		o.DBName,
		o.RepositoryCollection,
		o.BaseGroupCollection,
		o.ReviewerCollection)

	var session *mgo.Session
	var err error
	if o.UseSSL {
		logger.Info("Using SSL for mongodb connection...")
		uri, err := StripSSLFromURI(o.MongoURI)
		if err != nil {
			logger.Errorf("failed to strip ssl from mongo connection url: %v", err)
			return nil, err
		}

		dialInfo, err := mgo.ParseURL(uri)
		if err != nil {
			logger.Errorf("failed to parse URI: %v", err)
			return nil, err
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), &tls.Config{})
			if err != nil {
				logger.Errorf("failed to dial mongo: %v", err)
			}
			return conn, err
		}

		session, err = mgo.DialWithInfo(dialInfo)
		if err != nil {
			logger.Errorf("failed to connect to mongodb using ssl: %v", err)
			return nil, err
		}

	} else {
		session, err = mgo.Dial(o.MongoURI)
		if err != nil {
			logger.Errorf("failed to connect to mongodb: %v", err)
			return nil, err
		}
	}

	return &MongoStore{
		sesson:  session,
		Options: o,
	}, nil
}

func (ms *MongoStore) getCollection(collection string) (*mgo.Session, *mgo.Collection) {
	session := ms.sesson.Copy()
	col := session.DB(ms.Options.DBName).C(collection)
	return session, col
}

func (ms *MongoStore) PopLRUReviewer(ctx context.Context, alias []string) (*types.Reviewer, error) {
	// Lock, trying to fake a transaction using mongo
	// TODO: Create an actual concurrency solution i.e: stop using Cosmos mongo driver :)
	ms.reviewerWriteLock.Lock()
	defer ms.reviewerWriteLock.Unlock()

	lruReviewer, err := ms.GetLRUReviewer(ctx, alias)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	lruReviewer.LastReviewTime = time.Now().UTC()

	if err := ms.UpdateReviewer(ctx, lruReviewer); err != nil {
		return nil, errors.WithStack(err)
	}

	return lruReviewer, nil
}

func (ms *MongoStore) GetLRUReviewer(ctx context.Context, alias []string) (*types.Reviewer, error) {
	session, col := ms.getCollection(ms.Options.ReviewerCollection)
	defer session.Close()

	var reviewer types.Reviewer
	err := col.Find(bson.M{"alias": bson.M{"$in": alias}}).Sort("lastreviewtime").One(&reviewer)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.WithStack(ErrNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &reviewer, nil
}

func (ms *MongoStore) AddReviewer(ctx context.Context, reviewer *types.Reviewer) error {
	session, col := ms.getCollection(ms.Options.ReviewerCollection)
	defer session.Close()

	if err := col.Insert(reviewer); err != nil {
		return err
	}

	return nil
}

func (ms *MongoStore) UpdateReviewer(ctx context.Context, reviewer *types.Reviewer) error {
	session, col := ms.getCollection(ms.Options.ReviewerCollection)
	defer session.Close()

	if err := col.UpdateId(reviewer.Id, reviewer); err != nil {
		return err
	}

	return nil
}

func (ms *MongoStore) GetReviewerByADOID(ctx context.Context, adoID string) (*types.Reviewer, error) {
	session, col := ms.getCollection(ms.Options.ReviewerCollection)
	defer session.Close()

	var reviewer types.Reviewer
	err := col.Find(bson.M{"id": adoID}).One(&reviewer)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.WithStack(ErrNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &reviewer, nil
}

func (ms *MongoStore) GetReviewer(ctx context.Context, alias string) (*types.Reviewer, error) {
	session, col := ms.getCollection(ms.Options.ReviewerCollection)
	defer session.Close()

	var reviewer types.Reviewer
	err := col.Find(bson.M{"alias": alias}).One(&reviewer)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.WithStack(ErrNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &reviewer, nil
}

// AddRepository adds a repository to the mongo database
func (ms *MongoStore) AddRepository(ctx context.Context, repo *types.Repository) error {
	session, col := ms.getCollection(ms.Options.RepositoryCollection)
	defer session.Close()

	if err := col.Insert(repo); err != nil {
		return err
	}

	return nil
}

// UpdateRepository updates a repository in the database
func (ms *MongoStore) UpdateRepository(ctx context.Context, id string, repository *types.Repository) error {
	session, col := ms.getCollection(ms.Options.RepositoryCollection)
	defer session.Close()

	bsonID := bson.ObjectIdHex(id)

	if err := col.UpdateId(bsonID, repository); err != nil {
		return fmt.Errorf("MongoStore.UpdateRepository: %v", err)
	}

	return nil
}

// DeleteRepository deletes a reposotory from the database
func (ms *MongoStore) DeleteRepository(ctx context.Context, id string) error {
	session, col := ms.getCollection(ms.Options.RepositoryCollection)
	defer session.Close()

	bsonID := bson.ObjectIdHex(id)

	if err := col.RemoveId(bsonID); err != nil {
		return err
	}

	return nil
}

// GetRepositoryByID retrieves a repository by it's AdoID
func (ms *MongoStore) GetRepositoryByID(ctx context.Context, id string) (*types.Repository, error) {
	session, col := ms.getCollection(ms.Options.RepositoryCollection)
	defer session.Close()

	bsonID := bson.ObjectIdHex(id)

	var repo types.Repository
	err := col.Find(bson.M{"_id": bsonID}).One(&repo)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.WithStack(ErrNotFound)
		}
		return nil, err
	}

	return &repo, nil
}

// GetAllRepositories Retrives all repositories
func (ms *MongoStore) GetAllRepositories(ctx context.Context) ([]*types.Repository, error) {
	session, col := ms.getCollection(ms.Options.RepositoryCollection)
	defer session.Close()

	var repos []*types.Repository
	err := col.Find(nil).All(&repos)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if repos == nil {
		return []*types.Repository{}, nil
	}

	return repos, nil
}

// GetRepositoryByName gets a repository by it's name
func (ms *MongoStore) GetRepositoryByName(ctx context.Context, name, project string) (*types.Repository, error) {
	session, col := ms.getCollection(ms.Options.RepositoryCollection)
	defer session.Close()

	var repo types.Repository
	err := col.Find(bson.M{"name": name, "projectName": project}).One(&repo)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.WithStack(ErrNotFound)
		}
		return nil, err
	}

	return &repo, nil
}

// DeleteBaseGroup deletes a base group from the database
func (ms *MongoStore) DeleteBaseGroup(ctx context.Context, id string) error {
	session, col := ms.getCollection(ms.Options.BaseGroupCollection)
	defer session.Close()

	bsonID := bson.ObjectIdHex(id)

	if err := col.RemoveId(bsonID); err != nil {
		return err
	}

	return nil
}

// StripSSLFromURI removes the ssl from an URI
func StripSSLFromURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	q, _ := url.ParseQuery(u.RawQuery)
	q.Del("ssl")
	u.RawQuery = q.Encode()

	return u.String(), nil
}
