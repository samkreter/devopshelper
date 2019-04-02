package store

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/globalsign/mgo"
	"gopkg.in/mgo.v2/bson"
)

const (
	reviewerGroupCollectionName = "reviewerGroups"
)

type ReviewerStore interface {
	GetReviewerGroups()
	UpdateReviewerGroups()
	AddReviewerGroups()
}

type MongoStoreOptions struct {
	UseSSL   bool
	MongoURI string
	DBName   string
}

type MongoStore struct {
	ReviewerGroupCollectionName string
	sesson                      *mgo.Session
	Options                     *MongoStoreOptions
}

func (ms *MongoStore) Close() {
	ms.sesson.Close()
}

func NewMongoStore(o *MongoStoreOptions) (*MongoStore, error) {
	var session *mgo.Session
	var err error
	if o.UseSSL {
		log.Println("Using SSL for mongodb connection...")
		uri, err := StripSSLFromURI(o.MongoURI)
		if err != nil {
			log.Println("failed to strip ssl from mongo connection url: %v", err)
			return nil, err
		}

		dialInfo, err := mgo.ParseURL(uri)
		if err != nil {
			log.Println("failed to parse URI: %v", err)
			return nil, err
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), &tls.Config{})
			if err != nil {
				log.Println("failed to dial mongo: %v", err)
			}
			return conn, err
		}

		session, err = mgo.DialWithInfo(dialInfo)
		if err != nil {
			log.Println("failed to connect to mongodb using ssl: %v", err)
			return nil, err
		}

	} else {
		session, err = mgo.Dial(o.MongoURI)
		if err != nil {
			log.Println("failed to connect to mongodb: %v", err)
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

func (ms *MongoStore) test() {
	session, c := ms.getCollection("testCollection")
	defer session.Close()

	//Record container started the work
	err := c.Insert(&State{
		Name:  containerName,
		State: "InProgress",
	})
	if err != nil {
		log.Fatal(err)
	}

	var states []State
	err = c.Find(bson.M{}).All(&states)
	if err != nil {
		log.Fatal(err)
	}

	output := fmt.Sprintf("Hello World!\nOriginal String: %s\n\nEncrypted String:\n%x\n", work, t)

	// Record finish work in the database
	c.Update(bson.M{"name": containerName}, bson.M{"$set": bson.M{"state": "Done", "output": output}})

	time.Sleep(time.Minute * 5)
}

func dowork(work string) {
	hash := md5.New()

	t := []byte(work)

	for i := 0; i < 100; i++ {
		t = hash.Sum(t)
		fmt.Printf("%x\n", t)
	}

	time.Sleep(time.Second * 1)

	for i := 0; i < 100; i++ {
		t = hash.Sum(t)
		fmt.Printf("%x\n", t)
	}
}

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
