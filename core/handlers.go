package core

// ICommandHandler responds to text messages started from forwardslash
type ICommandHandler interface {
	HandleCommand(*Message, IBot) error
}

// IDocumentHandler responds to documents sent in chah
type IDocumentHandler interface {
	HandleDocument(*Document, IBot) error
}

// ITextHandler responds to all the text messages
type ITextHandler interface {
	HandleText(*Message, IBot) error
}

// IImageHandler responds to images
type IImageHandler interface {
	HandleImage(*Image, *Message, IBot) error
}
