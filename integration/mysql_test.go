package integration_test

import (
	"context"
	"database/sql"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo specs use dot imports by convention.
	. "github.com/onsi/gomega"    //nolint:revive // Ginkgo specs use dot imports by convention.
	"github.com/testcontainers/testcontainers-go"

	"github.com/akfaiz/migris/internal/testutil"
	"github.com/akfaiz/migris/schema"
)

var _ = Describe("MySQL", Ordered, func() {
	var (
		mCtx    context.Context
		db      *sql.DB
		tc      testcontainers.Container
		builder schema.Builder
	)

	BeforeAll(func() {
		var err error
		mCtx = context.Background()
		tc, db, err = testutil.StartMySQLTestDB(mCtx)
		Expect(err).NotTo(HaveOccurred())
		builder, err = schema.NewBuilder("mysql")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		_ = db.Close()
		if tc != nil {
			_ = tc.Terminate(mCtx)
		}
	})

	// dropAll removes all tables from the test database. MySQL DDL auto-commits,
	// so transactions cannot roll back table creation — explicit cleanup is required.
	dropAll := func() {
		tx, err := db.BeginTx(mCtx, nil)
		if err != nil {
			return
		}
		c := schema.NewContext(mCtx, tx)
		tables, err := builder.GetTables(c)
		if err != nil {
			_ = tx.Rollback()
			return
		}
		for _, t := range tables {
			_ = builder.DropIfExists(c, t.Name)
		}
		_ = tx.Commit()
	}

	newCtx := func() (schema.Context, *sql.Tx) {
		tx, err := db.BeginTx(mCtx, nil)
		Expect(err).NotTo(HaveOccurred())
		return schema.NewContext(mCtx, tx), tx
	}

	AfterEach(func() {
		dropAll()
	})

	Describe("Create", func() {
		It("returns error for nil context", func() {
			Expect(builder.Create(nil, "t", func(_ *schema.Blueprint) {})).To(HaveOccurred())
		})

		It("returns error for empty table name", func() {
			c, tx := newCtx()
			defer tx.Rollback()
			Expect(builder.Create(c, "", func(_ *schema.Blueprint) {})).To(HaveOccurred())
		})

		It("returns error for nil blueprint", func() {
			c, tx := newCtx()
			defer tx.Rollback()
			Expect(builder.Create(c, "t", nil)).To(HaveOccurred())
		})

		It("creates table with common column types", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
				t.String("email", 255).Unique()
				t.String("password", 255).Nullable()
				t.Boolean("active").Default(true)
				t.Timestamps()
			})).To(Succeed())

			has, err := builder.HasTable(c, "users")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("creates table with InnoDB engine", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "products", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
				t.InnoDB()
			})).To(Succeed())

			has, err := builder.HasTable(c, "products")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("creates table with charset and collation", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
				t.Charset("utf8mb4")
				t.Collation("utf8mb4_unicode_ci")
			})).To(Succeed())
		})

		It("creates table with composite primary key", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "user_roles", func(t *schema.Blueprint) {
				t.Integer("user_id")
				t.Integer("role_id")
				t.Primary("user_id", "role_id")
			})).To(Succeed())
		})

		It("creates table with foreign key and CASCADE actions", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
			})).To(Succeed())

			Expect(builder.Create(c, "orders", func(t *schema.Blueprint) {
				t.ID()
				t.UnsignedBigInteger("user_id")
				t.Decimal("amount", 10, 2)
				t.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("CASCADE")
			})).To(Succeed())
		})

		It("creates table with SoftDeletes", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "posts", func(t *schema.Blueprint) {
				t.ID()
				t.String("title", 255)
				t.SoftDeletes("deleted_at")
			})).To(Succeed())

			has, err := builder.HasColumn(c, "posts", "deleted_at")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("creates table with NumericMorphs", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "comments", func(t *schema.Blueprint) {
				t.ID()
				t.NumericMorphs("commentable")
				t.Text("body")
			})).To(Succeed())

			hasC, errC := builder.HasColumns(c, "comments", []string{"commentable_type", "commentable_id"})
			Expect(errC).NotTo(HaveOccurred())
			Expect(hasC).To(BeTrue())
		})

		It("creates table with UUIDMorphs", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "taggables", func(t *schema.Blueprint) {
				t.ID()
				t.UUIDMorphs("taggable")
			})).To(Succeed())

			cols, err := builder.GetColumns(c, "taggables")
			Expect(err).NotTo(HaveOccurred())
			Expect(colNames(cols)).To(ContainElements("taggable_type", "taggable_id"))
		})

		It("creates table with ForeignID.Constrained", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
			})).To(Succeed())

			Expect(builder.Create(c, "posts", func(t *schema.Blueprint) {
				t.ID()
				t.String("title", 255)
				t.ForeignID("user_id").Constrained()
			})).To(Succeed())

			has, err := builder.HasColumn(c, "posts", "user_id")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("creates table with RawColumn definition", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "events", func(t *schema.Blueprint) {
				t.ID()
				t.RawColumn("payload", "MEDIUMBLOB")
				t.String("source", 100)
			})).To(Succeed())

			has, err := builder.HasColumn(c, "events", "payload")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("creates TEMPORARY table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "tmp_import", func(t *schema.Blueprint) {
				t.ID()
				t.String("payload", 500)
				t.Temporary()
			})).To(Succeed())
		})

		It("creates table with AutoIncrementStartingValues", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "invoices", func(t *schema.Blueprint) {
				t.ID()
				t.String("ref", 50)
				t.AutoIncrementStartingValues(1000)
			})).To(Succeed())
		})

		It("creates table with Datetimes helper", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "logs", func(t *schema.Blueprint) {
				t.ID()
				t.Text("message")
				t.Datetimes()
			})).To(Succeed())

			hasC, errC := builder.HasColumns(c, "logs", []string{"created_at", "updated_at"})
			Expect(errC).NotTo(HaveOccurred())
			Expect(hasC).To(BeTrue())
		})

		It("returns error for duplicate table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
			})).To(Succeed())

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
			})).To(HaveOccurred())
		})
	})

	Describe("Table (Alter)", func() {
		It("adds columns to existing table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.String("phone", 20).Nullable()
			})).To(Succeed())

			has, err := builder.HasColumn(c, "users", "phone")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("adds column with After positioning", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.String("name", 255).After("id")
			})).To(Succeed())

			has, err := builder.HasColumn(c, "users", "name")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("adds invisible column", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.String("internal_code", 50).Nullable().Invisible()
			})).To(Succeed())
		})

		It("adds column with charset and collation modifiers", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.String("bio", 1000).Nullable().Charset("utf8mb4").Collation("utf8mb4_unicode_ci")
			})).To(Succeed())
		})

		It("adds Timestamp column with OnUpdate", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.Timestamp("updated_at").Nullable().UseCurrentOnUpdate()
			})).To(Succeed())
		})

		It("drops SoftDeletes column via DropSoftDeletes", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "articles", func(t *schema.Blueprint) {
				t.ID()
				t.String("title", 255)
				t.SoftDeletes("deleted_at")
			})).To(Succeed())

			Expect(builder.Table(c, "articles", func(t *schema.Blueprint) {
				t.DropSoftDeletes()
			})).To(Succeed())

			has, err := builder.HasColumn(c, "articles", "deleted_at")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
		})

		It("drops column", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
				t.String("bio", 1000).Nullable()
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.DropColumn("bio")
			})).To(Succeed())

			has, err := builder.HasColumn(c, "users", "bio")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
		})

		It("renames column", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("full_name", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.RenameColumn("full_name", "name")
			})).To(Succeed())

			has, err := builder.HasColumn(c, "users", "name")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("adds and drops index", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.Index("email").Name("idx_users_email")
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.DropIndex("idx_users_email")
			})).To(Succeed())
		})

		It("creates SpatialIndex on geometry column", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "locations", func(t *schema.Blueprint) {
				t.ID()
				t.Geometry("coords", "", 0)
			})).To(Succeed())

			Expect(builder.Table(c, "locations", func(t *schema.Blueprint) {
				t.SpatialIndex("coords")
			})).To(Succeed())
		})
	})

	Describe("Drop / DropIfExists / Rename", func() {
		It("drops existing table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "tmp", func(t *schema.Blueprint) { t.ID() })).To(Succeed())
			Expect(builder.Drop(c, "tmp")).To(Succeed())

			has, err := builder.HasTable(c, "tmp")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
		})

		It("returns error when dropping non-existent table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Drop(c, "no_such_table")).To(HaveOccurred())
		})

		It("DropIfExists does not error for missing table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.DropIfExists(c, "never_existed")).To(Succeed())
		})

		It("renames an existing table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "old_name", func(t *schema.Blueprint) { t.ID() })).To(Succeed())
			Expect(builder.Rename(c, "old_name", "new_name")).To(Succeed())

			has, err := builder.HasTable(c, "new_name")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})
	})

	Describe("Introspection", func() {
		It("HasTable, HasColumn, HasColumns, HasIndex work correctly", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255).Unique("uk_users_email")
				t.String("name", 255).Nullable()
			})).To(Succeed())

			has, err := builder.HasTable(c, "users")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
			has, err = builder.HasTable(c, "no_table")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
			has, err = builder.HasColumn(c, "users", "email")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
			has, err = builder.HasColumn(c, "users", "phone")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
			has, err = builder.HasColumns(c, "users", []string{"email", "name"})
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
			has, err = builder.HasIndex(c, "users", []string{"email"})
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("GetColumns returns column metadata", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
				t.String("email", 255).Unique()
				t.String("password", 255).Nullable()
				t.Timestamps()
			})).To(Succeed())

			cols, err := builder.GetColumns(c, "users")
			Expect(err).NotTo(HaveOccurred())
			Expect(cols).To(HaveLen(6))
			Expect(colNames(cols)).To(ContainElements("id", "name", "email", "password", "created_at", "updated_at"))
		})

		It("GetTables lists all tables", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "table_a", func(t *schema.Blueprint) { t.ID() })).To(Succeed())
			Expect(builder.Create(c, "table_b", func(t *schema.Blueprint) { t.ID() })).To(Succeed())

			tables, err := builder.GetTables(c)
			Expect(err).NotTo(HaveOccurred())
			names := make([]string, len(tables))
			for i, tb := range tables {
				names[i] = tb.Name
			}
			Expect(names).To(ContainElements("table_a", "table_b"))
		})
	})
})
