package blueprint

const (
	CommandAdd                         string = "add"
	CommandCreate                      string = "create"
	CommandChange                      string = "change"
	CommandDrop                        string = "drop"
	CommandDropIfExists                string = "dropIfExists"
	CommandRename                      string = "rename"
	CommandDropColumn                  string = "dropColumn"
	CommandRenameColumn                string = "renameColumn"
	CommandIndex                       string = "index"
	CommandUnique                      string = "unique"
	CommandPrimary                     string = "primary"
	CommandFullText                    string = "fulltext"
	CommandSpatialIndex                string = "spatialIndex"
	CommandVectorIndex                 string = "vectorIndex"
	CommandDropIndex                   string = "dropIndex"
	CommandDropUnique                  string = "dropUnique"
	CommandDropPrimary                 string = "dropPrimary"
	CommandDropFullText                string = "dropFulltext"
	CommandDropSpatialIndex            string = "dropSpatialIndex"
	CommandRenameIndex                 string = "renameIndex"
	CommandForeign                     string = "foreign"
	CommandDropForeign                 string = "dropForeign"
	CommandTableComment                string = "tableComment"
	CommandAutoIncrementStartingValues string = "autoIncrementStartingValues"
)

// Command represents a database command to be executed in a blueprint.
type Command struct {
	Name               string
	To                 string
	From               string
	Column             *Column
	Columns            []string
	Algorithm          string
	OperatorClass      string
	On                 string
	OnDelete           string
	OnUpdate           string
	References         []string
	Index              string
	Comment            string
	Value              int
	Deferrable         *bool
	InitiallyImmediate *bool
	Language           string
}
