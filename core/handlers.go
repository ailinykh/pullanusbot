package core

type ICommandHandler interface {
	HandleCommand(*Message, IBot) error
}

type IDocumentHandler interface {
	HandleDocument(*Document, IBot) error
}

type ITextHandler interface {
	HandleText(*Message, IBot) error
}

type IImageHandler interface {
	HandleImage(*Image, *Message, IBot) error
}
