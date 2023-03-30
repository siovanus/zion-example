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

package boot

import (
	"github.com/ethereum/go-ethereum/consensus/hotstuff/backend"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/polynetwork/zion-example/modules/cfg"
	"github.com/polynetwork/zion-example/modules/cross_chain_manager"
	"github.com/polynetwork/zion-example/modules/economic"
	"github.com/polynetwork/zion-example/modules/info_sync"
	"github.com/polynetwork/zion-example/modules/node_manager"
	"github.com/polynetwork/zion-example/modules/proposal_manager"
	"github.com/polynetwork/zion-example/modules/side_chain_manager"
)

func init() {
	//set genesis state of module contract when init genesis
	core.RegGenesis = node_manager.SetupGenesis
}

func InitModuleContracts() {
	//register module contract address map
	params.RegisterModuleContractAddrMap(cfg.ModuleNodeManager, cfg.NodeManagerContractAddress)
	params.RegisterModuleContractAddrMap(cfg.ModuleEconomic, cfg.EconomicContractAddress)
	params.RegisterModuleContractAddrMap(cfg.ModuleInfoSync, cfg.InfoSyncContractAddress)
	params.RegisterModuleContractAddrMap(cfg.ModuleCrossChain, cfg.CrossChainManagerContractAddress)
	params.RegisterModuleContractAddrMap(cfg.ModuleSideChainManager, cfg.SideChainManagerContractAddress)
	params.RegisterModuleContractAddrMap(cfg.ModuleProposalManager, cfg.ProposalManagerContractAddress)

	//set coinbase address to recieve accumulated gas fee
	params.CoinBaseAddress = cfg.NodeManagerContractAddress
	//set get validator method for consensus
	backend.GetGovernanceInfo = node_manager.GetGovernanceInfo

	//if these module have system tx, they will be executed at this order each end of the block
	economic.InitEconomic()
	node_manager.InitNodeManager()
	info_sync.InitInfoSync()
	cross_chain_manager.InitCrossChainManager()
	side_chain_manager.InitSideChainManager()
	proposal_manager.InitProposalManager()

	log.Info("Initialize module contracts",
		"node manager", cfg.NodeManagerContractAddress.Hex(),
		"economic", cfg.EconomicContractAddress.Hex(),
		"header sync", cfg.InfoSyncContractAddress.Hex(),
		"cross chain manager", cfg.CrossChainManagerContractAddress.Hex(),
		"side chain manager", cfg.SideChainManagerContractAddress.Hex(),
		"proposal manager", cfg.ProposalManagerContractAddress.Hex(),
	)
}
