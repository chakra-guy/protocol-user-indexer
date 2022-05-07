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

type EventIndexer struct {
	ID               int
	LastBlockIndexed uint64
	Spec             EventIndexerSpec
}

type EventIndexerSpec struct {
	Condition struct {
		Contract struct {
			Address string
			ABI     string
		}
		Event struct {
			Name string
		}
	}
	User struct {
		Event struct {
			Arg string
		}
	}
}
