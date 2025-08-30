package schema

const (
	commandAdd          string = "add"
	commandChange       string = "change"
	commandCreate       string = "create"
	commandDrop         string = "drop"
	commandDropIfExists string = "dropIfExists"
	commandDropColumn   string = "dropColumn"
	commandDropForeign  string = "dropForeign"
	commandDropFullText string = "dropFullText"
	commandDropIndex    string = "dropIndex"
	commandDropPrimary  string = "dropPrimary"
	commandDropUnique   string = "dropUnique"
	commandForeign      string = "foreign"
	commandFullText     string = "fullText"
	commandIndex        string = "index"
	commandPrimary      string = "primary"
	commandRename       string = "rename"
	commandRenameColumn string = "renameColumn"
	commandRenameIndex  string = "renameIndex"
	commandUnique       string = "unique"
)

type command struct {
	column             *columnDefinition
	deferrable         *bool
	initiallyImmediate *bool
	algorithm          string
	from               string
	index              string
	language           string
	name               string
	on                 string
	onDelete           string
	onUpdate           string
	to                 string
	columns            []string
	references         []string
}
