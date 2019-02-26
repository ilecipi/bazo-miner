package miner

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"reflect"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

//We can't use polymorphism, e.g. we can't use tx.verify() because the Transaction interface doesn't declare
//the verify method. This is because verification depends on the State (e.g., dynamic properties), which
//should only be of concern to the miner, not to the protocol package. However, this has the disadvantage
//that we have to do case distinction here.
func verify(tx protocol.Transaction) bool {
	var verified bool

	switch tx.(type) {
	case *protocol.FundsTx:
		verified = verifyFundsTx(tx.(*protocol.FundsTx))
	case *protocol.AccTx:
		verified = verifyAccTx(tx.(*protocol.AccTx))
	case *protocol.ConfigTx:
		verified = verifyConfigTx(tx.(*protocol.ConfigTx))
	case *protocol.StakeTx:
		verified = verifyStakeTx(tx.(*protocol.StakeTx))
	case *protocol.AggTx:
		verified = verifyAggTx(tx.(*protocol.AggTx))
	}

	return verified
}

func verifyFundsTx(tx *protocol.FundsTx) bool {
	if tx == nil {
		return false
	}

	//fundsTx only makes sense if amount > 0
	if tx.Amount == 0 || tx.Amount > MAX_MONEY {
		logger.Printf("Invalid transaction amount: %v\n", tx.Amount)
		return false
	}
	fmt.Println(tx)
	//Check if accounts are present in the actual state
	//accFrom := storage.State[tx.From]
	//accTo := storage.State[tx.To]

	//Accounts non existent
	//if accFrom == nil || accTo == nil {
	//	logger.Printf("Account non existent. From: %v\nTo: %v\n", accFrom, accTo)
	//	return false
	//}
	//accFromHash := protocol.SerializeHashContent(accFrom.Address)
	//accToHash := protocol.SerializeHashContent(accTo.Address)

	txHash := tx.Hash()

	pubKey := crypto.GetPubKeyFromAddressED(tx.From)
	validation :=ed25519.Verify(pubKey, txHash[:], tx.Sig[:])
	fmt.Println(validation)
	if ed25519.Verify(pubKey, txHash[:], tx.Sig[:]) && !reflect.DeepEqual(tx.From, tx.To) {
		return true
	} else {
		logger.Printf("Sig invalid. FromHash: %x\nToHash: %x\n", tx.From[0:8], tx.To[0:8])
		FileConnectionsLog.WriteString(fmt.Sprintf("Sig invalid. FromHash: %x\nToHash: %x\n", tx.From[0:8], tx.To[0:8]))
		return false
	}
}

func verifyAccTx(tx *protocol.AccTx) bool {
	if tx == nil {
		return false
	}

	for _, rootAcc := range storage.RootKeys {

		pubKey := crypto.GetPubKeyFromAddressED(rootAcc.Address)
		txHash := tx.Hash()

		//Only the hash of the pubkey is hashed and verified here
		if ed25519.Verify(pubKey, txHash[:], tx.Sig[:]) == true {
			return true
		}
	}

	return false
}

func verifyConfigTx(tx *protocol.ConfigTx) bool {
	if tx == nil {
		return false
	}

	//account creation can only be done with a valid priv/pub key which is hard-coded
	r, s := new(big.Int), new(big.Int)

	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	for _, rootAcc := range storage.RootKeys {
		pubKey := crypto.GetPubKeyFromAddressED(rootAcc.Address)
		txHash := tx.Hash()
		if ed25519.Verify(pubKey, txHash[:], tx.Sig[:]) == true {
			return true
		}
	}

	return false
}

func verifyStakeTx(tx *protocol.StakeTx) bool {
	if tx == nil {
		logger.Println("Transactions does not exist.")
		return false
	}

	//Check if account is present in the actual state
	acc := storage.State[tx.Account]
	if acc == nil {
		// TODO: Requires a Mutex?
		newAcc := protocol.NewAccount(tx.Account, [32]byte{}, 0, false, [crypto.COMM_KEY_LENGTH_ED]byte{}, nil, nil)
		acc = &newAcc
		storage.WriteAccount(acc)
	}

	r, s := new(big.Int), new(big.Int)

	r.SetBytes(tx.Sig[:32])
	s.SetBytes(tx.Sig[32:])

	tx.Account = acc.Address

	txHash := tx.Hash()

	pubKey := crypto.GetPubKeyFromAddressED(acc.Address)

	return ed25519.Verify(pubKey, txHash[:], tx.Sig[:])
}

func verifyAggTx(tx *protocol.AggTx) bool {
	if tx == nil {
		logger.Println("Transactions does not exist.")
		return false
	}

	//Check if accounts are existent
	//accSender, err := storage.GetAccount(tx.From)
	//if tx.From //!= protocol.SerializeHashContent(accSender.Address) || tx.To == nil || err != nil {
	//	logger.Printf("Account non existent. From: %v\nTo: %v\n%v", tx.From, tx.To, err)
	//	return false
	//}

	return true
}

//Returns true if id is in the list of possible ids and rational value for payload parameter.
//Some values just don't make any sense and have to be restricted accordingly
func parameterBoundsChecking(id uint8, payload uint64) bool {
	switch id {
	case protocol.BLOCK_SIZE_ID:
		if payload >= protocol.MIN_BLOCK_SIZE && payload <= protocol.MAX_BLOCK_SIZE {
			return true
		}
	case protocol.DIFF_INTERVAL_ID:
		if payload >= protocol.MIN_DIFF_INTERVAL && payload <= protocol.MAX_DIFF_INTERVAL {
			return true
		}
	case protocol.FEE_MINIMUM_ID:
		if payload >= protocol.MIN_FEE_MINIMUM && payload <= protocol.MAX_FEE_MINIMUM {
			return true
		}
	case protocol.BLOCK_INTERVAL_ID:
		if payload >= protocol.MIN_BLOCK_INTERVAL && payload <= protocol.MAX_BLOCK_INTERVAL {
			return true
		}
	case protocol.BLOCK_REWARD_ID:
		if payload >= protocol.MIN_BLOCK_REWARD && payload <= protocol.MAX_BLOCK_REWARD {
			return true
		}
	case protocol.STAKING_MINIMUM_ID:
		if payload >= protocol.MIN_STAKING_MINIMUM && payload <= protocol.MAX_STAKING_MINIMUM {
			return true
		}
	case protocol.WAITING_MINIMUM_ID:
		if payload >= protocol.MIN_WAITING_TIME && payload <= protocol.MAX_WAITING_TIME {
			return true
		}
	case protocol.ACCEPTANCE_TIME_DIFF_ID:
		if payload >= protocol.MIN_ACCEPTANCE_TIME_DIFF && payload <= protocol.MAX_ACCEPTANCE_TIME_DIFF {
			return true
		}
	case protocol.SLASHING_WINDOW_SIZE_ID:
		if payload >= protocol.MIN_SLASHING_WINDOW_SIZE && payload <= protocol.MAX_SLASHING_WINDOW_SIZE {
			return true
		}
	case protocol.SLASHING_REWARD_ID:
		if payload >= protocol.MIN_SLASHING_REWARD && payload <= protocol.MAX_SLASHING_REWARD {
			return true
		}
	}

	return false
}
