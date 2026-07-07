package store

import (
	"drug_trials_tool/internal/models"
)

// TrialInstitutionLink represents one trial-institution relationship.
type TrialInstitutionLink struct {
	TrialID          string `json:"trial_id"`
	InstitutionID    int    `json:"institution_id"`
	InvestigatorName string `json:"investigator_name"`
}

// ListAllTrialInstitutions returns all trial-institution links.
func (s *Store) ListAllTrialInstitutions() ([]TrialInstitutionLink, error) {
	rows, err := s.db.Query(`
		SELECT trial_id, institution_id, COALESCE(investigator_name,'')
		FROM trial_institutions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TrialInstitutionLink
	for rows.Next() {
		var l TrialInstitutionLink
		if err := rows.Scan(&l.TrialID, &l.InstitutionID, &l.InvestigatorName); err != nil {
			return nil, err
		}
		result = append(result, l)
	}
	return result, nil
}

// CountProvinces returns the number of provinces in the table.
func (s *Store) CountProvinces() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM provinces`).Scan(&count)
	return count, err
}

// ListProvinces2 returns all provinces, handling NULL code column.
func (s *Store) ListProvinces2() ([]models.Province, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.name, COALESCE(p.code,''),
			(SELECT COUNT(DISTINCT ti.trial_id)
			 FROM trial_institutions ti
			 JOIN institutions i ON i.id = ti.institution_id
			 WHERE i.province_id = p.id)
		FROM provinces p ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Province
	for rows.Next() {
		var p models.Province
		if err := rows.Scan(&p.ID, &p.Name, &p.Code, &p.TrialCount); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}
