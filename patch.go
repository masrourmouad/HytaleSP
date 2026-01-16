package main

import (
	"fmt"

	"github.com/itchio/headway/state"
	"github.com/itchio/lake/pools/fspool"
	"github.com/itchio/savior/filesource"
	"github.com/itchio/wharf/pwr/bowl"
	"github.com/itchio/wharf/pwr/patcher"

	_ "github.com/itchio/wharf/decompressors/cbrotli"
)


func applyPatch(source string, destination string, patchFilename string) {

	consumer := &state.Consumer {
		OnMessage: func(level string, message string) {
			fmt.Printf("[%s] %s\n", level, message);
		},
	}


	patchReader, _ := filesource.Open(patchFilename);
	p, _ := patcher.New(patchReader, consumer);

	targetPool := fspool.New(p.GetTargetContainer(), source);

	b, _ := bowl.NewFreshBowl(bowl.FreshBowlParams{
		SourceContainer: p.GetSourceContainer(),
		TargetContainer: p.GetTargetContainer(),
		TargetPool: targetPool,
		OutputFolder: destination,
	});

	// start the patch
	p.Resume(nil, targetPool, b);

}
