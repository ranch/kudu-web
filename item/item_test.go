package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"golang.org/x/net/context"

	"github.com/knq/firebase"
	pb "github.com/rnd/kudu-proto/item"
	pt "github.com/rnd/kudu-proto/types"
)

var testServer *server
var defaultContext context.Context
var defaultCancel context.CancelFunc
var userId string

func mockData() {
	var req pb.AddRequest

	testData := []pt.Item{
		{
			Goal:  "Foo",
			Tag:   "Bar",
			Notes: "# Baz",
			Url:   "brank.as",
		},
		{
			Goal:  "Kudu",
			Tag:   "App",
			Notes: "## Test",
			Url:   "google.com",
		},
	}

	for _, test := range testData {
		req.Item = &test
		_, err := testServer.AddItem(defaultContext, &req)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func clearData() {
	var err error

	path := fmt.Sprintf(itemRef, userId)

	keys := make(map[string]interface{})
	err = testServer.itemRef.Ref(path).Get(&keys, firebase.Shallow)
	if err != nil {
		log.Fatal(err)
	}

	for key := range keys {
		err = testServer.itemRef.Ref(path + key).Remove()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestMain(m *testing.M) {
	testServer = newServer()

	//TODO: Fix this by fetch test user id.
	userId = "foo"
	defaultContext, defaultCancel = context.WithCancel(context.WithValue(context.Background(), "userid", userId))

	clearData()
	mockData()

	code := m.Run()

	os.Exit(code)
}

func TestListItem(t *testing.T) {
	var err error

	test := pb.ListRequest{
		Goal: "Foo",
		Tag:  "Bar",
	}
	res, err := testServer.ListItem(defaultContext, &test)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Items) != 2 {
		t.Errorf("Expected list to have 2 items, got: %d", len(res.Items))
	}
}

func TestAddItem(t *testing.T) {
	var err error
	var req pb.AddRequest

	req.Item = &pt.Item{
		Goal:  "Foo",
		Tag:   "Bar",
		Notes: "# Baz",
		Url:   "reddit.com",
	}
	res, err := testServer.AddItem(defaultContext, &req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Id == "" {
		t.Error("Expected id to not empty")
	}
}

func TestGetItem(t *testing.T) {
	var err error

	keys := make(map[string]interface{})
	err = testServer.itemRef.Ref("/item").Get(&keys, firebase.Shallow)
	if err != nil {
		t.Fatal("Failed to get item keys")
	}

	for key := range keys {
		res, err := testServer.GetItem(defaultContext, &pb.GetRequest{Id: key})
		if err != nil {
			t.Errorf("Got error on get item with key: %s, %v", key, err)
		}
		if res.Item == nil {
			t.Error("Expected response item to not empty")
		}
	}
}
