package core

type ICommandHandler interface {
	HandleCommand(string, IBot) error
}

type IDocumentHandler interface {
	HandleDocument(*Document, IBot) error
}

type ITextHandler interface {
	HandleText(string, *User, IBot) error
}

type IImageHandler interface {
	HandleImage(*File, IBot) error
}
