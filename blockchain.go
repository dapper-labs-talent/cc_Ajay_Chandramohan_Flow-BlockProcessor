package main

import (
	"fmt"
	"sync"
)

type BlockProcessor struct {
	rwLock                 *sync.RWMutex
	finalBlockToHeight     map[string]uint64 // blockId to height
	processingBlockIdCount map[string]int64
}

func (p *BlockProcessor) checkIfValid(startHeight uint64, blocks []string) (bool, uint64) {
	processingHeight := uint64(len(p.finalBlockToHeight))
	if processingHeight < startHeight || processingHeight >= uint64(startHeight)+uint64(len(blocks)) {
		return false, uint64(len(p.finalBlockToHeight)) - 1
	}
	return true, uint64(len(p.finalBlockToHeight)) - 1
}

func (p *BlockProcessor) ProcessBlocks(startHeight uint64, blocks []string) uint64 {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()
	isValid, currentHeight := p.checkIfValid(startHeight, blocks)
	if !isValid {
		return currentHeight
	}
	processingHeight := uint64(len(p.finalBlockToHeight))
	relvantBlockId := blocks[processingHeight-startHeight]
	p.processingBlockIdCount[relvantBlockId]++
	if p.processingBlockIdCount[relvantBlockId] == 3 {
		_, ok := p.finalBlockToHeight[relvantBlockId]
		// check if the block has already been accepted at previous height
		if ok {
			return uint64(len(p.finalBlockToHeight)) - 1
		}
		p.finalBlockToHeight[relvantBlockId] = uint64(len(p.finalBlockToHeight))
		p.processingBlockIdCount = make(map[string]int64)
	}
	return uint64(len(p.finalBlockToHeight)) - 1
}

func (p *BlockProcessor) print() {
	fmt.Println(fmt.Sprintf("finalBlocks: %v  processingBlocks: %v", p.finalBlockToHeight, p.processingBlockIdCount))
}
