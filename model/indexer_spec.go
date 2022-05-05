package model

type TxIndexer struct {
	ID               int
	LastIndexedBlock int
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
