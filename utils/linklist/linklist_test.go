package linklist

import (
    "os"

	"testing"
)

const testFileName = "test_filename"

func TestUniq(t *testing.T) {
    testString := "www.example.com"
    if !Uniq(testString) {
        t.Fatalf("uniq on empty map returns false")
    }
    if Uniq(testString) {
        t.Fatalf("double key returns true")
    }
}

func TestSave(t *testing.T) {
    llist.Save(testFileName)
    if _, err := os.Stat(testFileName); os.IsNotExist(err) {
        t.Fatal("Did not create the file whan calling save")
    } else {
        os.Remove(testFileName)
    }
}
