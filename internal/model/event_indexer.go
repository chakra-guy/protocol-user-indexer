package model

type EventIndexer struct {
	ID               int
	LastBlockIndexed int
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
