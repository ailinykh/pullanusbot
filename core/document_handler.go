package core

type IDocumentHandler interface {
	HandleDocument(*Document, IBot) error
}
