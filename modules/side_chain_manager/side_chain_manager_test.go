/*
 * Copyright (C) 2021 The Zion Authors
 * This file is part of The Zion library.
 *
 * The Zion is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The Zion is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The Zion.  If not, see <http://www.gnu.org/licenses/>.
 */

package side_chain_manager

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contract"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/polynetwork/zion-example/modules/cfg"
	"github.com/polynetwork/zion-example/modules/go_abi/side_chain_manager_abi"
	"github.com/polynetwork/zion-example/modules/node_manager"
	"github.com/stretchr/testify/assert"
)

func init() {
	InitSideChainManager()
	node_manager.InitNodeManager()
	db := rawdb.NewMemoryDatabase()
	sdb, _ = state.New(common.Hash{}, state.NewDatabase(db), nil)
}

var (
	sdb     *state.StateDB
	signers []common.Address
)

func init() {
	node_manager.InitNodeManager()
	sdb = contract.NewTestStateDB()
	signers, _ = contract.GenerateTestPeers(2)
	node_manager.StoreGenesisEpoch(sdb, signers, signers)
}

func testRegisterSideChainManager(t *testing.T) {
	param := new(RegisterSideChainParam)
	param.ChainID = 8
	param.Name = "mychain"
	param.Router = 3

	input, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodRegisterSideChain, param)
	assert.Nil(t, err)
	param.Name = strings.Repeat("1", 100)
	param.ChainID = 9
	param.ExtraInfo = make([]byte, 1000000)
	param.CCMCAddress = make([]byte, 1000)
	input1, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodRegisterSideChain, param)
	assert.Nil(t, err)

	blockNumber := big.NewInt(1)
	extra := uint64(2100000000)
	tr := contract.NewTimer(side_chain_manager_abi.MethodRegisterSideChain)
	for _, input := range [][]byte{input, input1} {
		caller := signers[0]
		tr.Start()
		contractRef := contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		ret, leftOverGas, err := contractRef.ModuleCall(common.Address{}, cfg.SideChainManagerContractAddress, input)
		tr.Stop()
		assert.Nil(t, err)
		result, err := contract.PackOutputs(ABI, side_chain_manager_abi.MethodRegisterSideChain)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)

		contract := contract.NewModuleContract(sdb, contractRef)
		sideChain, err := GetSideChainApply(contract, 8)
		assert.Equal(t, sideChain.Name, "mychain")
		assert.Nil(t, err)

		_, _, err = contractRef.ModuleCall(common.Address{}, cfg.SideChainManagerContractAddress, input)
		assert.NotNil(t, err)
	}
	tr.Dump()
}

func testApproveRegisterSideChain(t *testing.T) {
	testRegisterSideChainManager(t)

	param := new(ChainIDParam)
	param.ChainID = 8

	input, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodApproveRegisterSideChain, param)
	assert.Nil(t, err)
	param.ChainID = 9
	input1, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodApproveRegisterSideChain, param)
	assert.Nil(t, err)

	extra := uint64(2100000000)
	tr := contract.NewTimer(side_chain_manager_abi.MethodApproveRegisterSideChain)
	for _, input := range [][]byte{input, input1} {
		caller := signers[0]
		blockNumber := big.NewInt(1)
		contractRef := contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		tr.Start()
		ret, leftOverGas, err := contractRef.ModuleCall(caller, cfg.SideChainManagerContractAddress, input)
		tr.Stop()
		assert.Nil(t, err)

		result, err := contract.PackOutputs(ABI, side_chain_manager_abi.MethodApproveRegisterSideChain, true)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)

		caller = signers[1]
		contractRef = contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		tr.Start()
		ret, leftOverGas, err = contractRef.ModuleCall(caller, cfg.SideChainManagerContractAddress, input)
		tr.Stop()
		assert.Nil(t, err)

		result, err = contract.PackOutputs(ABI, side_chain_manager_abi.MethodApproveRegisterSideChain, true)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)
	}
	tr.Dump()
}

func testUpdateSideChain(t *testing.T) {
	testApproveRegisterSideChain(t)

	param := new(RegisterSideChainParam)
	param.ChainID = 8
	param.Name = "own"
	param.Router = 3

	input, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodUpdateSideChain, param)
	assert.Nil(t, err)

	param.ChainID = 9
	param.Name = strings.Repeat("2", 100)
	param.ExtraInfo = make([]byte, 1000000)
	param.CCMCAddress = make([]byte, 1000)
	input1, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodUpdateSideChain, param)
	assert.Nil(t, err)

	extra := uint64(2100000000)
	tr := contract.NewTimer(side_chain_manager_abi.MethodUpdateSideChain)
	for _, input := range [][]byte{input, input1} {
		blockNumber := big.NewInt(1)
		caller := signers[0]
		contractRef := contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		tr.Start()
		ret, leftOverGas, err := contractRef.ModuleCall(common.Address{}, cfg.SideChainManagerContractAddress, input)
		tr.Stop()
		assert.Nil(t, err)

		result, err := contract.PackOutputs(ABI, side_chain_manager_abi.MethodUpdateSideChain)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)
	}
	tr.Dump()

}

func testApproveUpdateSideChain(t *testing.T) {
	testUpdateSideChain(t)
	param := new(ChainIDParam)
	param.ChainID = 8

	input, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodApproveUpdateSideChain, param)
	assert.Nil(t, err)

	param.ChainID = 9
	input1, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodApproveUpdateSideChain, param)
	assert.Nil(t, err)

	extra := uint64(2100000000)
	tr := contract.NewTimer(side_chain_manager_abi.MethodApproveUpdateSideChain)
	tr1 := contract.NewTimer(side_chain_manager_abi.MethodGetSideChain)
	for i, input := range [][]byte{input, input1} {
		blockNumber := big.NewInt(1)
		caller := signers[0]
		contractRef := contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		tr.Start()
		ret, leftOverGas, err := contractRef.ModuleCall(caller, cfg.SideChainManagerContractAddress, input)
		tr.Stop()
		assert.Nil(t, err)

		result, err := contract.PackOutputs(ABI, side_chain_manager_abi.MethodApproveUpdateSideChain, true)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)

		caller = signers[1]
		contractRef = contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		tr.Start()
		ret, leftOverGas, err = contractRef.ModuleCall(caller, cfg.SideChainManagerContractAddress, input)
		tr.Stop()
		assert.Nil(t, err)

		result, err = contract.PackOutputs(ABI, side_chain_manager_abi.MethodApproveUpdateSideChain, true)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)

		c := contract.NewModuleContract(sdb, contractRef)
		sideChain, err := GetSideChainObject(c, 8+uint64(i))
		assert.Nil(t, err)
		assert.NotNil(t, sideChain)
		if i == 0 {
			assert.Equal(t, sideChain.Name, "own")
			param.ChainID = 8
		} else {
			param.ChainID = 9
		}
		input, err = contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodGetSideChain, param)
		assert.Nil(t, err)

		contractRef = contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		tr1.Start()
		ret, leftOverGas, err = contractRef.ModuleCall(caller, cfg.SideChainManagerContractAddress, input)
		tr1.Stop()
		assert.Nil(t, err)
		result, err = contract.PackOutputs(ABI, side_chain_manager_abi.MethodGetSideChain, sideChain)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)
	}
	tr.Dump()
	tr1.Dump()
}

func testQuiteSideChain(t *testing.T) {
	testApproveUpdateSideChain(t)
	param := new(ChainIDParam)
	param.ChainID = 8

	input, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodQuitSideChain, param)
	assert.Nil(t, err)

	param.ChainID = 9
	input1, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodQuitSideChain, param)
	assert.Nil(t, err)

	extra := uint64(2100000000)
	tr := contract.NewTimer(side_chain_manager_abi.MethodQuitSideChain)
	for _, input := range [][]byte{input, input1} {
		blockNumber := big.NewInt(1)
		caller := signers[0]
		contractRef := contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		tr.Start()
		ret, leftOverGas, err := contractRef.ModuleCall(caller, cfg.SideChainManagerContractAddress, input)
		tr.Stop()
		assert.Nil(t, err)

		result, err := contract.PackOutputs(ABI, side_chain_manager_abi.MethodQuitSideChain)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)
	}
	tr.Dump()
}

func TestApproveQuiteSideChain(t *testing.T) {
	testQuiteSideChain(t)
	param := new(ChainIDParam)
	param.ChainID = 8

	input, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodApproveQuitSideChain, param)
	assert.Nil(t, err)

	param.ChainID = 9
	input1, err := contract.PackMethodWithStruct(ABI, side_chain_manager_abi.MethodApproveQuitSideChain, param)
	assert.Nil(t, err)

	extra := uint64(2100000000)
	tr := contract.NewTimer(side_chain_manager_abi.MethodApproveQuitSideChain)
	for i, input := range [][]byte{input, input1} {
		blockNumber := big.NewInt(1)
		caller := signers[0]
		contractRef := contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		tr.Start()
		ret, leftOverGas, err := contractRef.ModuleCall(caller, cfg.SideChainManagerContractAddress, input)
		tr.Stop()
		assert.Nil(t, err)

		result, err := contract.PackOutputs(ABI, side_chain_manager_abi.MethodApproveUpdateSideChain, true)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)

		caller = signers[1]
		contractRef = contract.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, extra, nil)
		tr.Start()
		ret, leftOverGas, err = contractRef.ModuleCall(caller, cfg.SideChainManagerContractAddress, input)
		tr.Stop()
		assert.Nil(t, err)

		result, err = contract.PackOutputs(ABI, side_chain_manager_abi.MethodApproveQuitSideChain, true)
		assert.Nil(t, err)
		assert.Equal(t, ret, result)
		assert.Equal(t, leftOverGas, extra)

		contract := contract.NewModuleContract(sdb, contractRef)
		sideChain, err := GetSideChainObject(contract, 8+uint64(i))
		assert.Nil(t, err)
		assert.Nil(t, sideChain)
	}
	tr.Dump()
}
