package main

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

func Assert(t *testing.T, expected uint64, actual uint64) {
	if expected != actual {
		t.Error(fmt.Sprintf("Expected %v Actual %v", expected, actual))
	}
}

func AssertBool(t *testing.T, expected bool, actual bool) {
	if expected != actual {
		t.Error(fmt.Sprintf("Expected %v Actual %v", expected, actual))
	}
}

func Test_Process_One_by_One_Block(t *testing.T) {
	blockProcessor := BlockProcessor{
		rwLock:                 &sync.RWMutex{},
		finalBlockToHeight:     map[string]uint64{"Genesis Block": uint64(0)},
		processingBlockIdCount: make(map[string]int64),
	}
	for i := 0; i < 2; i++ {
		Assert(t, 0, blockProcessor.ProcessBlocks(1, []string{"id-1"}))
	}
	blockProcessor.print()
	Assert(t, 1, blockProcessor.ProcessBlocks(1, []string{"id-1"}))
	blockProcessor.print()
}

func Test_Process_Multiple_BlockIds_Same_height(t *testing.T) {
	blockProcessor := BlockProcessor{
		rwLock:                 &sync.RWMutex{},
		finalBlockToHeight:     map[string]uint64{"Genesis Block": uint64(0)},
		processingBlockIdCount: make(map[string]int64),
	}
	for i := 0; i < 2; i++ {
		Assert(t, 0, blockProcessor.ProcessBlocks(1, []string{"id-1"}))
		Assert(t, 0, blockProcessor.ProcessBlocks(1, []string{"id-2"}))
	}
	blockProcessor.print()
	//id-1 wins as its first to reach 3 times
	Assert(t, 1, blockProcessor.ProcessBlocks(1, []string{"id-1"}))
	Assert(t, 1, blockProcessor.ProcessBlocks(1, []string{"id-2"}))

	//Lets try height 2
	for i := 0; i < 2; i++ {
		Assert(t, 1, blockProcessor.ProcessBlocks(2, []string{"id-2"}))
	}
	Assert(t, 2, blockProcessor.ProcessBlocks(2, []string{"id-2"}))
	blockProcessor.print()
}

func Test_Process_Skip_Duplicate_Blocks(t *testing.T) {
	blockProcessor := BlockProcessor{
		rwLock:                 &sync.RWMutex{},
		finalBlockToHeight:     map[string]uint64{"Genesis Block": uint64(0)},
		processingBlockIdCount: make(map[string]int64),
	}
	for i := 0; i < 2; i++ {
		Assert(t, 0, blockProcessor.ProcessBlocks(1, []string{"id-1"}))
	}
	blockProcessor.print()
	//id-1 is accepted as it reached 3 times
	Assert(t, 1, blockProcessor.ProcessBlocks(1, []string{"id-1"}))

	//Lets try height 2
	for i := 0; i < 2; i++ {
		Assert(t, 1, blockProcessor.ProcessBlocks(2, []string{"id-1"}))
	}
	//id-1 isnt accepted due to being duplicate
	Assert(t, 1, blockProcessor.ProcessBlocks(2, []string{"id-1"}))
	blockProcessor.print()
}

/*
If race condition it will throw error "fatal error: concurrent map writes", (in case you remove the locks)
20M requests can finish in 4.26 seconds in my machine, which makes it 4.6 millions per second !!!
I didnt see a point in optimizing further
*/
func Test_Concurrency_and_load_testing(t *testing.T) {
	blockProcessor := BlockProcessor{
		rwLock:                 &sync.RWMutex{},
		finalBlockToHeight:     map[string]uint64{"Genesis Block": uint64(0)},
		processingBlockIdCount: make(map[string]int64),
	}

	var wg sync.WaitGroup

	for i := 1; i <= 200; i++ {
		for height := 1; height <= 100; height++ {
			for id := 1; id <= 1000; id++ {
				wg.Add(1)

				go func(currentId int) {
					defer wg.Done()
					blockProcessor.ProcessBlocks(uint64(height), []string{"id-" + strconv.FormatInt(int64(currentId), 10)})
				}(id)
			}
		}
	}
	wg.Wait()
	blockProcessor.print()
}

func Test_isValid(t *testing.T) {
	blockProcessor := BlockProcessor{
		rwLock: &sync.RWMutex{},
		finalBlockToHeight: map[string]uint64{"Genesis Block": uint64(0),
			"blockid-1": uint64(1),
			"blockid-2": uint64(2),
		},
		processingBlockIdCount: make(map[string]int64),
	}
	{
		isValid, _ := blockProcessor.checkIfValid(0, []string{"block-1"})
		AssertBool(t, false, isValid)
	}

	{
		isValid, _ := blockProcessor.checkIfValid(1, []string{"block-1"})
		AssertBool(t, false, isValid)
	}

	{
		isValid, _ := blockProcessor.checkIfValid(1, []string{"block-1", "block-2"})
		AssertBool(t, false, isValid)
	}

	{
		isValid, _ := blockProcessor.checkIfValid(1, []string{"block-1", "block-2", "block-3"})
		AssertBool(t, true, isValid)
	}

	{
		isValid, _ := blockProcessor.checkIfValid(3, []string{"block-1"})
		AssertBool(t, true, isValid)
	}
}
