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

var _ = Describe("PostgreSQL", Ordered, func() {
	var (
		gCtx    context.Context
		db      *sql.DB
		tc      testcontainers.Container
		builder schema.Builder
	)

	BeforeAll(func() {
		var err error
		gCtx = context.Background()
		tc, db, err = testutil.StartPostgresTestDB(gCtx)
		Expect(err).NotTo(HaveOccurred())
		builder, err = schema.NewBuilder("postgres")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		_ = db.Close()
		if tc != nil {
			_ = tc.Terminate(gCtx)
		}
	})

	// newCtx opens a fresh transaction. Callers must defer tx.Rollback().
	// Postgres DDL is transactional, so rollback cleans up created tables.
	newCtx := func() (schema.Context, *sql.Tx) {
		tx, err := db.BeginTx(gCtx, nil)
		Expect(err).NotTo(HaveOccurred())
		return schema.NewContext(gCtx, tx), tx
	}

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
				t.BigInteger("user_id")
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

			cols, err := builder.GetColumns(c, "posts")
			Expect(err).NotTo(HaveOccurred())
			var deletedAtCol *schema.Column
			for _, col := range cols {
				if col.Name == "deleted_at" {
					deletedAtCol = col
					break
				}
			}
			Expect(deletedAtCol).NotTo(BeNil())
			Expect(deletedAtCol.Nullable).To(BeTrue())
		})

		It("creates table with SoftDeletesTz", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "events", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 255)
				t.SoftDeletesTz("deleted_at")
			})).To(Succeed())

			has, err := builder.HasColumn(c, "events", "deleted_at")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("creates table with NumericMorphs columns and index", func() {
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

		It("creates table with RawColumn", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "events", func(t *schema.Blueprint) {
				t.ID()
				t.RawColumn("metadata", "JSONB")
				t.String("name", 255)
			})).To(Succeed())

			has, err := builder.HasColumn(c, "events", "metadata")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
		})

		It("creates a TEMPORARY table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "tmp_import", func(t *schema.Blueprint) {
				t.ID()
				t.String("payload", 500)
				t.Temporary()
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

		It("creates table with NullableTimestamps", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "sessions", func(t *schema.Blueprint) {
				t.ID()
				t.String("token", 100)
				t.NullableTimestamps()
			})).To(Succeed())

			cols, err := builder.GetColumns(c, "sessions")
			Expect(err).NotTo(HaveOccurred())
			for _, col := range cols {
				if col.Name == "created_at" || col.Name == "updated_at" {
					Expect(col.Nullable).To(BeTrue(), "expected %s to be nullable", col.Name)
				}
			}
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

		// VectorIndex requires pgvector extension not available in standard postgres:16-alpine.
		PIt("creates table with VectorIndex (requires pgvector)", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "items", func(t *schema.Blueprint) {
				t.ID()
				t.Vector("embedding", 1536)
			})).To(Succeed())

			Expect(builder.Table(c, "items", func(t *schema.Blueprint) {
				t.VectorIndex("embedding")
			})).To(Succeed())
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
				t.String("address", 255).Nullable()
			})).To(Succeed())

			hasC, errC := builder.HasColumns(c, "users", []string{"phone", "address"})
			Expect(errC).NotTo(HaveOccurred())
			Expect(hasC).To(BeTrue())
		})

		It("adds SoftDeletes column via Table alter", func() {
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

		It("modifies existing column with Change()", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 100)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.String("name", 255).Nullable().Change()
			})).To(Succeed())
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

		It("adds, renames, and drops index", func() {
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
				t.RenameIndex("idx_users_email", "idx_users_contact")
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.DropIndex("idx_users_contact")
			})).To(Succeed())

			has, err := builder.HasIndex(c, "users", []string{"idx_users_contact"})
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
		})

		It("adds and drops foreign key", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "roles", func(t *schema.Blueprint) {
				t.ID()
				t.String("name", 100)
			})).To(Succeed())

			Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
				t.ID()
				t.String("email", 255)
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.Integer("role_id").Nullable()
				t.Foreign("role_id").References("id").On("roles").OnDelete("SET NULL")
			})).To(Succeed())

			Expect(builder.Table(c, "users", func(t *schema.Blueprint) {
				t.DropForeign([]string{"role_id"})
			})).To(Succeed())
		})
	})

	Describe("Drop", func() {
		It("drops an existing table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "tmp", func(t *schema.Blueprint) {
				t.ID()
			})).To(Succeed())

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
	})

	Describe("DropIfExists", func() {
		It("drops existing table without error", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "tmp", func(t *schema.Blueprint) {
				t.ID()
			})).To(Succeed())

			Expect(builder.DropIfExists(c, "tmp")).To(Succeed())
		})

		It("does not error when table does not exist", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.DropIfExists(c, "never_existed")).To(Succeed())
		})
	})

	Describe("Rename", func() {
		It("renames an existing table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Create(c, "old_name", func(t *schema.Blueprint) {
				t.ID()
			})).To(Succeed())

			Expect(builder.Rename(c, "old_name", "new_name")).To(Succeed())

			has, err := builder.HasTable(c, "new_name")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeTrue())
			has, err = builder.HasTable(c, "old_name")
			Expect(err).NotTo(HaveOccurred())
			Expect(has).To(BeFalse())
		})

		It("returns error for non-existent source table", func() {
			c, tx := newCtx()
			defer tx.Rollback()

			Expect(builder.Rename(c, "ghost", "phantom")).To(HaveOccurred())
		})
	})

	Describe("Introspection", func() {
		Describe("HasTable", func() {
			It("returns true for existing table", func() {
				c, tx := newCtx()
				defer tx.Rollback()

				Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
					t.ID()
				})).To(Succeed())

				has, err := builder.HasTable(c, "users")
				Expect(err).NotTo(HaveOccurred())
				Expect(has).To(BeTrue())
			})

			It("returns false for non-existent table", func() {
				c, tx := newCtx()
				defer tx.Rollback()

				has, err := builder.HasTable(c, "no_such_table")
				Expect(err).NotTo(HaveOccurred())
				Expect(has).To(BeFalse())
			})
		})

		Describe("HasColumn / HasColumns", func() {
			It("returns true for existing column", func() {
				c, tx := newCtx()
				defer tx.Rollback()

				Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
					t.ID()
					t.String("email", 255)
				})).To(Succeed())

				has, err := builder.HasColumn(c, "users", "email")
				Expect(err).NotTo(HaveOccurred())
				Expect(has).To(BeTrue())
			})

			It("returns false for non-existent column", func() {
				c, tx := newCtx()
				defer tx.Rollback()

				Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
					t.ID()
				})).To(Succeed())

				has, err := builder.HasColumn(c, "users", "phone")
				Expect(err).NotTo(HaveOccurred())
				Expect(has).To(BeFalse())
			})

			It("returns true when all columns exist", func() {
				c, tx := newCtx()
				defer tx.Rollback()

				Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
					t.ID()
					t.String("name", 255)
					t.String("email", 255)
				})).To(Succeed())

				has, err := builder.HasColumns(c, "users", []string{"name", "email"})
				Expect(err).NotTo(HaveOccurred())
				Expect(has).To(BeTrue())
			})
		})

		Describe("HasIndex", func() {
			It("detects index by column list", func() {
				c, tx := newCtx()
				defer tx.Rollback()

				Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
					t.ID()
					t.String("email", 255).Unique("uk_users_email")
				})).To(Succeed())

				has, err := builder.HasIndex(c, "users", []string{"email"})
				Expect(err).NotTo(HaveOccurred())
				Expect(has).To(BeTrue())
			})

			It("returns false for non-existent index", func() {
				c, tx := newCtx()
				defer tx.Rollback()

				Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
					t.ID()
					t.String("email", 255)
				})).To(Succeed())

				has, err := builder.HasIndex(c, "users", []string{"email"})
				Expect(err).NotTo(HaveOccurred())
				Expect(has).To(BeFalse())
			})
		})

		Describe("GetColumns", func() {
			It("returns all columns with metadata", func() {
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
				Expect(colNames(cols)).To(ContainElements(
					"id", "name", "email", "password", "created_at", "updated_at",
				))
			})

			It("returns empty slice for non-existent table", func() {
				c, tx := newCtx()
				defer tx.Rollback()

				cols, err := builder.GetColumns(c, "ghost")
				Expect(err).NotTo(HaveOccurred())
				Expect(cols).To(BeEmpty())
			})
		})

		Describe("GetIndexes", func() {
			It("returns all indexes including primary", func() {
				c, tx := newCtx()
				defer tx.Rollback()

				Expect(builder.Create(c, "users", func(t *schema.Blueprint) {
					t.ID()
					t.String("email", 255).Unique("uk_users_email")
					t.Index("email").Name("idx_users_email")
				})).To(Succeed())

				idxs, err := builder.GetIndexes(c, "users")
				Expect(err).NotTo(HaveOccurred())
				Expect(idxs).NotTo(BeEmpty())
				Expect(indexNames(idxs)).To(ContainElement("uk_users_email"))
			})
		})

		Describe("GetTables", func() {
			It("lists created tables", func() {
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
})
