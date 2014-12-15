// Copyright (c) 2014-2014 PPCD developers.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/kac-/umint"
	"github.com/mably/btcutil"
	"github.com/mably/ppcwallet/txstore"
)

func (w *Wallet) CreateCoinStake(fromTime int64) (err error) {

	// Get current block's height and hash.
	bs, err := w.chainSvr.BlockStamp()
	if err != nil {
		return
	}

	bits, err := w.chainSvr.CurrentTarget()
	if err != nil {
		return
	}

	params, err := w.chainSvr.Params()
	if err != nil {
		return
	}
	stakeMinAge := params.StakeMinAge

	nCredit := int64(0)
	fKernelFound := false

	eligibles, err := w.findEligibleOutputs(6, bs)
	for _, eligible := range eligibles {
		// TODO verify min age requirement
		var block *txstore.Block
		block, err = eligible.Block()
		if err != nil {
			return
		}
		if block.KernelStakeModifier == btcutil.KernelStakeModifierUnknown {
			var ksm uint64
			ksm, err = w.chainSvr.GetKernelStakeModifier(&block.Hash)
			if err != nil {
				log.Errorf("Error getting kernel stake modifier for block %v", &block.Hash)
				return
			} else {
				log.Infof("Found kernel stake modifier for block %v: %v", &block.Hash, ksm)
				block.KernelStakeModifier = ksm
				w.TxStore.MarkDirty()
			}
		}
		// TODO verify that block.KernelStakeModifier is defined
		tx := eligible.Tx()
		for n := int64(0); n < 60 && !fKernelFound; n++ {
			stpl := umint.StakeKernelTemplate{
				//BlockFromTime:  int64(utx.BlockTime),
				BlockFromTime: block.Time.Unix(),
				//StakeModifier:  utx.StakeModifier,
				StakeModifier: block.KernelStakeModifier,
				//PrevTxOffset:   utx.OffsetInBlock,
				PrevTxOffset: tx.Offset(),
				//PrevTxTime:     int64(utx.Time),
				PrevTxTime: tx.MsgTx().Time.Unix(),
				//PrevTxOutIndex: outPoint.Index,
				PrevTxOutIndex: eligible.OutputIndex,
				//PrevTxOutValue: int64(utx.Value),
				PrevTxOutValue: int64(eligible.Amount()),
				IsProtocolV03:  true,
				StakeMinAge:    stakeMinAge,
				Bits:           bits,
				TxTime:         fromTime - n,
			}
			var success bool
			_, success, err, _ = umint.CheckStakeKernelHash(&stpl)
			if err != nil {
				log.Errorf("Check kernel hash error: %v", err)
				return
			}
			if success {
				log.Infof("Valid kernel hash found!")
				// TODO create coinstake tx
				fKernelFound = true
				break
			}
		}
		if fKernelFound {
			break
		}
	}

	log.Infof("Valid kernel hash found: %v", fKernelFound)

	if nCredit == 0 {
		return
	}

	// TODO to be continued...

	return
}
