package sqlite

import (
	"fmt"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
)

// HeadingDAO persists note headings in the SQLite database.
type HeadingDAO struct {
	tx     Transaction
	logger util.Logger

	// Prepared SQL statements
	addHeadingStmt     *LazyStmt
	removeHeadingsStmt *LazyStmt
}

// NewHeadingDAO creates a new instance of a DAO working on the given database
// transaction.
func NewHeadingDAO(tx Transaction, logger util.Logger) *HeadingDAO {
	return &HeadingDAO{
		tx:     tx,
		logger: logger,

		// Add a new heading.
		addHeadingStmt: tx.PrepareLazy(`
			INSERT INTO headings (note_id, level, title, line)
			VALUES (?, ?, ?, ?)
		`),

		// Remove all the headings of a note.
		removeHeadingsStmt: tx.PrepareLazy(`
			DELETE FROM headings
			 WHERE note_id = ?
		`),
	}
}

// Add inserts all the headings of the given note.
func (d *HeadingDAO) Add(id core.NoteID, headings []core.Heading) error {
	for _, heading := range headings {
		_, err := d.addHeadingStmt.Exec(noteIDToSQL(id), heading.Level, heading.Text, heading.Line)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveAll removes all the headings of the given note.
func (d *HeadingDAO) RemoveAll(id core.NoteID) error {
	_, err := d.removeHeadingsStmt.Exec(noteIDToSQL(id))
	return err
}

// FindByNoteIDs returns the headings of the given notes, indexed by note ID
// and sorted by their position in the note.
func (d *HeadingDAO) FindByNoteIDs(ids []core.NoteID) (map[core.NoteID][]core.Heading, error) {
	headings := map[core.NoteID][]core.Heading{}
	if len(ids) == 0 {
		return headings, nil
	}

	rows, err := d.tx.Query(fmt.Sprintf(`
		SELECT note_id, level, title, line
		  FROM headings
		 WHERE note_id IN (%s)
		 ORDER BY note_id, line
	`, joinNoteIDs(ids, ",")))
	if err != nil {
		return headings, err
	}
	defer rows.Close()

	for rows.Next() {
		var noteID int64
		var heading core.Heading
		err = rows.Scan(&noteID, &heading.Level, &heading.Text, &heading.Line)
		if err != nil {
			return headings, err
		}
		id := core.NoteID(noteID)
		headings[id] = append(headings[id], heading)
	}

	return headings, rows.Err()
}
