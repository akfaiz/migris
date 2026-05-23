package integration_test

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo specs use dot imports by convention.
	. "github.com/onsi/gomega"    //nolint:revive // Ginkgo specs use dot imports by convention.

	"github.com/akfaiz/migris/schema"
)

var _ = Describe("SQLite", Ordered, func() {
	const dbFile = "test_integration_sqlite.db"

	var (
		sCtx    context.Context
		db      *sql.DB
		builder schema.Builder
	)

	BeforeAll(func() {
		var err error
		sCtx = context.Background()
		db, err = sql.Open("sqlite3", dbFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Ping()).To(Succeed())
		builder, err = schema.NewBuilder("sqlite3")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		_ = db.Close()
		_ = os.Remove(dbFile)
	})

	// dropAll removes all tables after each spec for isolation.
	dropAll := func() {
		tx, err := db.BeginTx(sCtx, nil)
		if err != nil {
			return
		}
		c := schema.NewContext(sCtx, tx)
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
		tx, err := db.BeginTx(sCtx, nil)
		Expect(err).NotTo(HaveOccurred())
		return schema.NewContext(sCtx, tx), tx
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

		It("creates table with RawColumn", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "events", func(t *schema.Blueprint) {
				t.ID()
				t.RawColumn("payload", "BLOB")
				t.String("source", 100)
			})).To(Succeed())

			has, err := builder.HasColumn(c, "events", "payload")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
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

		It("creates table with composite primary key", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "user_roles", func(t *schema.Blueprint) {
				t.Integer("user_id")
				t.Integer("role_id")
				t.Primary("user_id", "role_id")
			})).To(Succeed())
		})

		It("returns error for duplicate table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) { t.ID() })).To(Succeed())
			Expect(builder.Create(c, "users", func(t *schema.Blueprint) { t.ID() })).To(HaveOccurred())
		})
	})

	Describe("Table (Alter)", func() {
		It("adds new columns", func() {
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

		It("adds SoftDeletes to existing table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "articles", func(t *schema.Blueprint) {
				t.ID()
				t.String("title", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "articles", func(t *schema.Blueprint) {
				t.SoftDeletes("deleted_at")
			})).To(Succeed())

			has, err := builder.HasColumn(c, "articles", "deleted_at")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("adds and drops unique index", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.Unique("email").Name("uk_users_email")
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.DropIndex("uk_users_email")
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
			has, err = builder.HasTable(c, "old_name")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
		})
	})

	Describe("Introspection", func() {
		It("HasTable / HasColumn / HasColumns / HasIndex", func() {
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
			has, err = builder.HasColumns(c, "users", []string{"email", "name"})
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
			has, err = builder.HasIndex(c, "users", []string{"email"})
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("GetColumns returns all columns", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
				t.String("email", 255)
				t.Timestamps()
			})).To(Succeed())

			cols, err := builder.GetColumns(c, "users")
			Expect(err).NotTo(HaveOccurred())
			Expect(cols).To(HaveLen(5))
			Expect(colNames(cols)).To(ContainElements("id", "name", "email", "created_at", "updated_at"))
		})

		It("GetTables lists all tables", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "alpha", func(t *schema.Blueprint) { t.ID() })).To(Succeed())
			Expect(builder.Create(c, "beta", func(t *schema.Blueprint) { t.ID() })).To(Succeed())

			tables, err := builder.GetTables(c)
			Expect(err).NotTo(HaveOccurred())
			names := make([]string, len(tables))
			for i, tb := range tables {
				names[i] = tb.Name
			}
			Expect(names).To(ContainElements("alpha", "beta"))
		})
	})
})
