package types

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

type EventIndexerSpec struct {
	Condition struct {
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
