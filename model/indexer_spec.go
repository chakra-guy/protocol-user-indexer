package model

type TxIndexer struct {
	ID               int
	LastBlockIndexed uint64
	Spec             TxIndexerSpec
}

type TxIndexerSpec struct {
	Condition struct {
		Tx struct {
			To string
		}
	}
	User struct {
		Tx string
	}
}
