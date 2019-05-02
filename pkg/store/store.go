package store

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/samkreter/go-core/log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/samkreter/vstsautoreviewer/pkg/types"
)

const (
	defaultBaseGroupCollectionName  = "basegroups"
	defaultRepositoryCollectionName = "repositories"
)

var (
	// ErrNotFound the error is not found
	ErrNotFound = errors.New("record not found")
)

// RepositoryStore holds information for a repository
type RepositoryStore interface {
	// Repository Ops
	AddRepository(ctx context.Context, repo *types.Repository) error
	UpdateRepository(ctx context.Context, id string, repository *types.Repository) error
	DeleteRepository(ctx context.Context, id string) error
	GetRepositoryByID(ctx context.Context, id string) (*types.Repository, error)
	GetAllRepositories(ctx context.Context) ([]*types.Repository, error)
	GetRepositoryByName(ctx context.Context, name, project string) (*types.Repository, error)

	// Base Group Ops
	AddBaseGroup(ctx context.Context, name string, group *types.BaseGroup) error
	UpdateBaseGroup(ctx context.Context, id string, group *types.BaseGroup) error
	DeleteBaseGroup(ctx context.Context, id string) error
	GetBaseGroupByName(ctx context.Context, name string) (*types.BaseGroup, error)
	GetAllBaseGroups(ctx context.Context) ([]*types.BaseGroup, error)
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
}

// MongoStore implementation to interact with a mongo database
type MongoStore struct {
	ReviewerGroupCollectionName string
	sesson                      *mgo.Session
	Options                     *MongoStoreOptions
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

	if o.RepositoryCollection == "" {
		o.RepositoryCollection = defaultRepositoryCollectionName
	}

	if o.BaseGroupCollection == "" {
		o.BaseGroupCollection = defaultBaseGroupCollectionName
	}

	logger.Infof("MongoStore: Using DB: '%s' for mongo with RepoCollection: %s, BaseGroupCollection: %s",
		o.DBName,
		o.RepositoryCollection,
		o.BaseGroupCollection)

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

// GetRepositoryByID retrieves a repository by it's ID
func (ms *MongoStore) GetRepositoryByID(ctx context.Context, id string) (*types.Repository, error) {
	session, col := ms.getCollection(ms.Options.RepositoryCollection)
	defer session.Close()

	bsonID := bson.ObjectIdHex(id)

	var repo types.Repository
	err := col.Find(bson.M{"_id": bsonID}).One(&repo)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrNotFound
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
		return nil, err
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
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &repo, nil
}

// AddBaseGroup adds a base group into the database
func (ms *MongoStore) AddBaseGroup(ctx context.Context, name string, group *types.BaseGroup) error {
	session, col := ms.getCollection(ms.Options.BaseGroupCollection)
	defer session.Close()

	if err := col.Insert(group); err != nil {
		return err
	}

	return nil
}

// UpdateBaseGroup updates a base group in the database
func (ms *MongoStore) UpdateBaseGroup(ctx context.Context, id string, group *types.BaseGroup) error {
	session, col := ms.getCollection(ms.Options.BaseGroupCollection)
	defer session.Close()

	bsonID := bson.ObjectIdHex(id)

	if err := col.UpdateId(bsonID, group); err != nil {
		return err
	}

	return nil
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

// GetBaseGroupByName get a basegroup by name
func (ms *MongoStore) GetBaseGroupByName(ctx context.Context, name string) (*types.BaseGroup, error) {
	session, col := ms.getCollection(ms.Options.BaseGroupCollection)
	defer session.Close()

	var group types.BaseGroup
	err := col.Find(bson.M{"name": name}).One(&group)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &group, nil
}

// GetAllBaseGroups gets all available base groups
func (ms *MongoStore) GetAllBaseGroups(ctx context.Context) ([]*types.BaseGroup, error) {
	session, col := ms.getCollection(ms.Options.RepositoryCollection)
	defer session.Close()

	var groups []*types.BaseGroup
	err := col.Find(nil).All(&groups)
	if err != nil {
		return nil, err
	}

	if groups == nil {
		return []*types.BaseGroup{}, nil
	}

	return groups, nil
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
