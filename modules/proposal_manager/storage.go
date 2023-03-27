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

package proposal_manager

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/contract"
	"github.com/ethereum/go-ethereum/contract/utils"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/polynetwork/zion-example/modules/node_manager"
)

var ErrEof = errors.New("EOF")

const (
	SKP_PROPOSAL_ID             = "st_proposal_id"
	SKP_PROPOSAL                = "st_proposal"
	SKP_PROPOSAL_LIST           = "st_proposal_list"
	SKP_CONFIG_PROPOSAL_LIST    = "st_config_proposal_list"
	SKP_COMMUNITY_PROPOSAL_LIST = "st_community_proposal_list"
)

func getProposalID(s *contract.ModuleContract) (*big.Int, error) {
	proposalID := new(big.Int)
	key := proposalIDKey()
	store, err := get(s, key)
	if err == ErrEof {
		return proposalID, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getProposalID, get store error: %v", err)
	}
	return new(big.Int).SetBytes(store), nil
}

func setProposalID(s *contract.ModuleContract, proposalID *big.Int) error {
	key := proposalIDKey()
	err := set(s, key, proposalID.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func getProposalList(s *contract.ModuleContract) (*ProposalList, error) {
	proposalList := &ProposalList{
		make([]*big.Int, 0),
	}
	key := proposalListKey()
	store, err := get(s, key)
	if err == ErrEof {
		return proposalList, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getProposalList, get store error: %v", err)
	}
	if err := rlp.DecodeBytes(store, proposalList); err != nil {
		return nil, fmt.Errorf("getProposalList, deserialize proposal list error: %v", err)
	}
	return proposalList, nil
}

func setProposalList(s *contract.ModuleContract, proposalList *ProposalList) error {
	key := proposalListKey()
	store, err := rlp.EncodeToBytes(proposalList)
	if err != nil {
		return fmt.Errorf("setProposalList, serialize proposalList error: %v", err)
	}
	err = set(s, key, store)
	if err != nil {
		return err
	}
	return nil
}

func removeFromProposalList(s *contract.ModuleContract, ID *big.Int) error {
	proposalList, err := getProposalList(s)
	if err != nil {
		return fmt.Errorf("removeFromProposalList, getProposalList error: %v", err)
	}

	j := 0
	for _, proposalID := range proposalList.ProposalList {
		if proposalID.Cmp(ID) != 0 {
			proposalList.ProposalList[j] = proposalID
			j++
		}
	}
	proposalList.ProposalList = proposalList.ProposalList[:j]
	err = setProposalList(s, proposalList)
	if err != nil {
		return fmt.Errorf("removeFromProposalList, setProposalList error: %v", err)
	}
	return nil
}

func removeExpiredFromProposalList(s *contract.ModuleContract) error {
	communityInfo, err := node_manager.GetCommunityInfoImpl(s)
	if err != nil {
		return fmt.Errorf("removeExpiredFromProposalList, node_manager.GetCommunityInfoImpl error: %v", err)
	}

	proposalList, err := getProposalList(s)
	if err != nil {
		return fmt.Errorf("removeExpiredFromProposalList, getProposalList error: %v", err)
	}
	if len(proposalList.ProposalList) == 0 {
		return nil
	}

	j := 0
	for _, proposalID := range proposalList.ProposalList {
		proposal, err := getProposal(s, proposalID)
		if err != nil {
			return fmt.Errorf("removeExpiredFromProposalList, getProposal error: %v", err)
		}
		if proposal.EndHeight.Cmp(s.ContractRef().BlockHeight()) > 0 {
			proposalList.ProposalList[j] = proposalID
			j++
		} else {
			// transfer token to community pool
			err = utils.ModuleTransfer(s.StateDB(), this, communityInfo.CommunityAddress, proposal.Stake)
			if err != nil {
				return fmt.Errorf("removeExpiredFromProposalList, utils.ModuleTransfer error: %v", err)
			}
		}
	}
	proposalList.ProposalList = proposalList.ProposalList[:j]
	err = setProposalList(s, proposalList)
	if err != nil {
		return fmt.Errorf("removeExpiredFromProposalList, setProposalList error: %v", err)
	}
	return nil
}

func getConfigProposalList(s *contract.ModuleContract) (*ConfigProposalList, error) {
	configProposalList := &ConfigProposalList{
		make([]*big.Int, 0),
	}
	key := configProposalListKey()
	store, err := get(s, key)
	if err == ErrEof {
		return configProposalList, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getConfigProposalList, get store error: %v", err)
	}
	if err := rlp.DecodeBytes(store, configProposalList); err != nil {
		return nil, fmt.Errorf("getConfigProposalList, deserialize config proposal list error: %v", err)
	}
	return configProposalList, nil
}

func setConfigProposalList(s *contract.ModuleContract, configProposalList *ConfigProposalList) error {
	key := configProposalListKey()
	store, err := rlp.EncodeToBytes(configProposalList)
	if err != nil {
		return fmt.Errorf("setConfigProposalList, serialize config proposal list error: %v", err)
	}
	err = set(s, key, store)
	if err != nil {
		return err
	}
	return nil
}

func cleanConfigProposalList(s *contract.ModuleContract) error {
	err := setConfigProposalList(s, &ConfigProposalList{make([]*big.Int, 0)})
	if err != nil {
		return fmt.Errorf("cleanConfigProposalList, setConfigProposalList error: %v", err)
	}
	return nil
}

func removeExpiredFromConfigProposalList(s *contract.ModuleContract) error {
	communityInfo, err := node_manager.GetCommunityInfoImpl(s)
	if err != nil {
		return fmt.Errorf("removeExpiredFromConfigProposalList, node_manager.GetCommunityInfoImpl error: %v", err)
	}

	configProposalList, err := getConfigProposalList(s)
	if err != nil {
		return fmt.Errorf("removeExpiredFromConfigProposalList, getProposalList error: %v", err)
	}
	if len(configProposalList.ConfigProposalList) == 0 {
		return nil
	}

	j := 0
	for _, proposalID := range configProposalList.ConfigProposalList {
		proposal, err := getProposal(s, proposalID)
		if err != nil {
			return fmt.Errorf("removeExpiredFromConfigProposalList, getProposal error: %v", err)
		}
		if proposal.EndHeight.Cmp(s.ContractRef().BlockHeight()) > 0 {
			configProposalList.ConfigProposalList[j] = proposalID
			j++
		} else {
			// transfer token to community pool
			err = utils.ModuleTransfer(s.StateDB(), this, communityInfo.CommunityAddress, proposal.Stake)
			if err != nil {
				return fmt.Errorf("removeExpiredFromConfigProposalList, utils.ModuleTransfer error: %v", err)
			}
		}
	}
	configProposalList.ConfigProposalList = configProposalList.ConfigProposalList[:j]
	err = setConfigProposalList(s, configProposalList)
	if err != nil {
		return fmt.Errorf("removeExpiredFromConfigProposalList, setProposalList error: %v", err)
	}
	return nil
}

func getCommunityProposalList(s *contract.ModuleContract) (*CommunityProposalList, error) {
	communityProposalList := &CommunityProposalList{
		make([]*big.Int, 0),
	}
	key := communityProposalListKey()
	store, err := get(s, key)
	if err == ErrEof {
		return communityProposalList, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getCommunityProposalList, get store error: %v", err)
	}
	if err := rlp.DecodeBytes(store, communityProposalList); err != nil {
		return nil, fmt.Errorf("getCommunityProposalList, deserialize community proposal list error: %v", err)
	}
	return communityProposalList, nil
}

func setCommunityProposalList(s *contract.ModuleContract, communityProposalList *CommunityProposalList) error {
	key := communityProposalListKey()
	store, err := rlp.EncodeToBytes(communityProposalList)
	if err != nil {
		return fmt.Errorf("setCommunityProposalList, serialize community proposal list error: %v", err)
	}
	err = set(s, key, store)
	if err != nil {
		return err
	}
	return nil
}

func cleanCommunityProposalList(s *contract.ModuleContract) error {
	err := setCommunityProposalList(s, &CommunityProposalList{make([]*big.Int, 0)})
	if err != nil {
		return fmt.Errorf("cleanCommunityProposalList, setCommunityProposalList error: %v", err)
	}
	return nil
}

func removeExpiredFromCommunityProposalList(s *contract.ModuleContract) error {
	communityInfo, err := node_manager.GetCommunityInfoImpl(s)
	if err != nil {
		return fmt.Errorf("removeExpiredFromCommunityProposalList, node_manager.GetCommunityInfoImpl error: %v", err)
	}

	communityProposalList, err := getCommunityProposalList(s)
	if err != nil {
		return fmt.Errorf("removeExpiredFromCommunityProposalList, getCommunityProposalList error: %v", err)
	}
	if len(communityProposalList.CommunityProposalList) == 0 {
		return nil
	}

	j := 0
	for _, proposalID := range communityProposalList.CommunityProposalList {
		proposal, err := getProposal(s, proposalID)
		if err != nil {
			return fmt.Errorf("removeExpiredFromConfigProposalList, getProposal error: %v", err)
		}
		if proposal.EndHeight.Cmp(s.ContractRef().BlockHeight()) > 0 {
			communityProposalList.CommunityProposalList[j] = proposalID
			j++
		} else {
			// transfer token to community pool
			err = utils.ModuleTransfer(s.StateDB(), this, communityInfo.CommunityAddress, proposal.Stake)
			if err != nil {
				return fmt.Errorf("removeExpiredFromCommunityProposalList, utils.ModuleTransfer error: %v", err)
			}
		}
	}
	communityProposalList.CommunityProposalList = communityProposalList.CommunityProposalList[:j]
	err = setCommunityProposalList(s, communityProposalList)
	if err != nil {
		return fmt.Errorf("removeExpiredFromCommunityProposalList, setCommunityProposalList error: %v", err)
	}
	return nil
}

func getProposal(s *contract.ModuleContract, ID *big.Int) (*Proposal, error) {
	proposal := new(Proposal)
	key := proposalKey(ID)
	store, err := get(s, key)
	if err != nil {
		return nil, fmt.Errorf("getProposal, get store error: %v", err)
	}
	if err := rlp.DecodeBytes(store, proposal); err != nil {
		return nil, fmt.Errorf("getProposal, deserialize proposal error: %v", err)
	}
	return proposal, nil
}

func setProposal(s *contract.ModuleContract, proposal *Proposal) error {
	key := proposalKey(proposal.ID)
	store, err := rlp.EncodeToBytes(proposal)
	if err != nil {
		return fmt.Errorf("setProposal, serialize proposal error: %v", err)
	}
	err = set(s, key, store)
	if err != nil {
		return err
	}
	return nil
}

// ====================================================================
//
// storage basic operations
//
// ====================================================================

func get(s *contract.ModuleContract, key []byte) ([]byte, error) {
	return customGet(s.GetCacheDB(), key)
}

func set(s *contract.ModuleContract, key, value []byte) error {
	err := customSet(s.GetCacheDB(), key, value)
	if err != nil {
		return err
	}
	return nil
}

func del(s *contract.ModuleContract, key []byte) error {
	err := customDel(s.GetCacheDB(), key)
	if err != nil {
		return err
	}
	return nil
}

func customGet(s *state.Store, key []byte) ([]byte, error) {
	value, err := s.Get(key)
	if err != nil {
		return nil, err
	} else if value == nil || len(value) == 0 {
		return nil, ErrEof
	} else {
		return value, nil
	}
}

func customSet(s *state.Store, key, value []byte) error {
	err := s.Put(key, value)
	if err != nil {
		return err
	}
	return nil
}

func customDel(s *state.Store, key []byte) error {
	err := s.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

// ====================================================================
//
// storage keys
//
// ====================================================================

func proposalIDKey() []byte {
	return utils.ConcatKey(this, []byte(SKP_PROPOSAL_ID))
}

func proposalKey(ID *big.Int) []byte {
	return utils.ConcatKey(this, []byte(SKP_PROPOSAL), ID.Bytes())
}

func proposalListKey() []byte {
	return utils.ConcatKey(this, []byte(SKP_PROPOSAL_LIST))
}

func configProposalListKey() []byte {
	return utils.ConcatKey(this, []byte(SKP_CONFIG_PROPOSAL_LIST))
}

func communityProposalListKey() []byte {
	return utils.ConcatKey(this, []byte(SKP_COMMUNITY_PROPOSAL_LIST))
}
