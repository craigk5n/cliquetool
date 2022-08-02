package main

import (
	"testing"
)

func TestReadParameterFile(t *testing.T) {
	testfile := "testdata/test_params.txt"
	groupParams, err := readParamFile(testfile)
	if err != nil {
		t.Errorf("Error reading file %s", testfile)
	}
	if groupParams.filename != testfile {
		t.Errorf("NodeDb node 0 external name wrong")
	}
	//for i := 0; i < len(groupParams.params); i++ {
	//	fmt.Printf("#%d - Group size %d minLinks %d\n", i, groupParams.params[i].groupSize, groupParams.params[i].minLinks)
	//}
	var paramTest = []struct {
		groupSize     int
		expectedLinks int
	}{
		{2, 1},
		{3, 2},
		{4, 3},
		{9, 6},
		{10, 7},
		{13, 7},
		{16, 8},
	}
	for i := 0; i < len(paramTest); i++ {
		want := paramTest[i].expectedLinks
		got := minLinksForGroupSize(&groupParams, paramTest[i].groupSize)
		if want != got {
			t.Errorf("minLinksForGroupSize wanted %d, got %d for size %d", want, got, paramTest[i].groupSize)
		}
	}
}

// Same test as above but with a different parameter file that uses
// decimal min link numbers instead of density fractions.
func TestReadParameterFile2(t *testing.T) {
	testfile := "testdata/test_params2.txt"
	groupParams, err := readParamFile(testfile)
	if err != nil {
		t.Errorf("Error reading file %s", testfile)
	}
	if groupParams.filename != testfile {
		t.Errorf("NodeDb node 0 external name wrong")
	}
	//for i := 0; i < len(groupParams.params); i++ {
	//	fmt.Printf("#%d - Group size %d minLinks %d\n", i, groupParams.params[i].groupSize, groupParams.params[i].minLinks)
	//}
	var paramTest = []struct {
		groupSize     int
		expectedLinks int
	}{
		{2, 1},
		{3, 2},
		{4, 3},
		{9, 6},
		{10, 7},
		{13, 7},
		{16, 8},
	}
	for i := 0; i < len(paramTest); i++ {
		want := paramTest[i].expectedLinks
		got := minLinksForGroupSize(&groupParams, paramTest[i].groupSize)
		if want != got {
			t.Errorf("minLinksForGroupSize wanted %d, got %d for size %d", want, got, paramTest[i].groupSize)
		}
	}
}

func TestNoParameterFile(t *testing.T) {
	var paramTest = []struct {
		groupSize     int
		expectedLinks int
	}{
		{2, 1},
		{3, 2},
		{4, 3},
		{9, 8},
		{10, 9},
		{13, 12},
		{16, 15},
	}
	for i := 0; i < len(paramTest); i++ {
		want := paramTest[i].expectedLinks
		got := minLinksForGroupSize(nil, paramTest[i].groupSize)
		if want != got {
			t.Errorf("minLinksForGroupSize wanted %d, got %d for size %d", want, got, paramTest[i].groupSize)
		}
	}
}
