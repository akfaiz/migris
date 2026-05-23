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

var _ = Describe("MariaDB", Ordered, func() {
	var (
		mbCtx   context.Context
		db      *sql.DB
		tc      testcontainers.Container
		builder schema.Builder
	)

	BeforeAll(func() {
		var err error
		mbCtx = context.Background()
		tc, db, err = testutil.StartMariaDBTestDB(mbCtx)
		Expect(err).NotTo(HaveOccurred())
		builder, err = schema.NewBuilder("mariadb")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		_ = db.Close()
		if tc != nil {
			_ = tc.Terminate(mbCtx)
		}
	})

	dropAll := func() {
		tx, err := db.BeginTx(mbCtx, nil)
		if err != nil {
			return
		}
		c := schema.NewContext(mbCtx, tx)
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
		tx, err := db.BeginTx(mbCtx, nil)
		Expect(err).NotTo(HaveOccurred())
		return schema.NewContext(mbCtx, tx), tx
	}

	AfterEach(func() {
		dropAll()
	})

	Describe("Create", func() {
		It("creates table with common column types", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
				t.String("email", 255).Unique()
				t.String("password", 255).Nullable()
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

		It("creates table with RawColumn", func() {
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

		It("creates table with AutoIncrementStartingValues", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "invoices", func(t *schema.Blueprint) {
				t.ID()
				t.String("ref", 50)
				t.AutoIncrementStartingValues(1000)
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
		It("adds columns", func() {
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
		})

		It("drops SoftDeletes column", func() {
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
	})

	Describe("Drop / DropIfExists / Rename", func() {
		It("drops table and DropIfExists are idempotent", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "tmp", func(t *schema.Blueprint) { t.ID() })).To(Succeed())
			Expect(builder.Drop(c, "tmp")).To(Succeed())
			Expect(builder.DropIfExists(c, "tmp")).To(Succeed())
		})

		It("renames table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "src", func(t *schema.Blueprint) { t.ID() })).To(Succeed())
			Expect(builder.Rename(c, "src", "dst")).To(Succeed())
			has, err := builder.HasTable(c, "dst")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})
	})

	Describe("Introspection", func() {
		It("GetColumns, GetTables, HasIndex work correctly", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255).Unique("uk_users_email")
				t.String("name", 255)
				t.Timestamps()
			})).To(Succeed())

			cols, err := builder.GetColumns(c, "users")
			Expect(err).NotTo(HaveOccurred())
			Expect(cols).To(HaveLen(5))

			has, err := builder.HasIndex(c, "users", []string{"email"})
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())

			tables, err := builder.GetTables(c)
			Expect(err).NotTo(HaveOccurred())
			names := make([]string, len(tables))
			for i, tb := range tables {
				names[i] = tb.Name
			}
			Expect(names).To(ContainElement("users"))
		})
	})
})
