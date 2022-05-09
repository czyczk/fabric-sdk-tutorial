package bcao

type IScrewBCAO interface {
	Transfer(sourceCorpName string, targetCorpName string, transferAmnt uint, eventID string) (string, error)
	Query(targetCorpName string) (string, error)
}
