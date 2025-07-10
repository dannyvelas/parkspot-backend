package storage

import (
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/storage/selectopts"
	"github.com/dannyvelas/parkspot-backend/util"
)

type ResidentRepoMock struct {
	residents []models.Resident
}

func NewResidentRepoMock() ResidentRepoMock {
	return ResidentRepoMock{residents: make([]models.Resident, 0)}
}

func (residentRepoMock *ResidentRepoMock) SelectWhere(residentFields models.Resident, selectOpts ...selectopts.SelectOpt) ([]models.Resident, error) {
	var residentsFound []models.Resident
	for _, resident := range residentRepoMock.residents {
		if (residentFields.ID == "" || residentFields.ID == resident.ID) &&
			(residentFields.FirstName == "" || residentFields.FirstName == resident.FirstName) &&
			(residentFields.LastName == "" || residentFields.LastName == resident.LastName) &&
			(residentFields.Phone == "" || residentFields.Phone == resident.Phone) &&
			(residentFields.Email == "" || residentFields.Email == resident.Email) {
			residentsFound = append(residentsFound, resident)
		}
	}
	return residentsFound, nil
}

func (residentRepoMock *ResidentRepoMock) SelectCountWhere(residentFields models.Resident, selectOpts ...selectopts.SelectOpt) (int, error) {
	return len(residentRepoMock.residents), nil
}

func (residentRepoMock *ResidentRepoMock) AddToAmtParkingDaysUsed(id string, days int) error {
	i := util.Find(residentRepoMock.residents, func(resident models.Resident) bool {
		return resident.ID == id
	})
	if i == -1 {
		return errs.NewNotFound("resident")
	}
	resident := &residentRepoMock.residents[i]

	*resident.AmtParkingDaysUsed = *resident.AmtParkingDaysUsed + days
	return nil
}

func (residentRepoMock *ResidentRepoMock) Create(resident models.Resident) error {
	// cast *resident.UnlimDays to bool
	if resident.UnlimDays == nil {
		resident.UnlimDays = util.ToPtr(false)
	}

	residentRepoMock.residents = append(residentRepoMock.residents, resident)

	return nil
}

func (residentRepoMock *ResidentRepoMock) Delete(id string) error {
	i := util.Find(residentRepoMock.residents, func(resident models.Resident) bool {
		return resident.ID == id
	})
	if i == -1 {
		return errs.NewNotFound("resident")
	}
	residents := residentRepoMock.residents

	// replace the element at the index you want to delete with the last element
	residents[i] = residents[len(residents)-1]

	// re-size slice to remove the last element
	residentRepoMock.residents = residents[:len(residents)-1]

	return nil
}

func (residentRepoMock *ResidentRepoMock) Update(residentFields models.Resident) error {
	i := util.Find(residentRepoMock.residents, func(resident models.Resident) bool { return resident.ID == residentFields.ID })
	if i < 0 {
		return errs.NewNotFound("resident")
	}
	resident := &residentRepoMock.residents[i]
	if residentFields.FirstName != "" {
		resident.FirstName = residentFields.FirstName
	}
	if residentFields.LastName != "" {
		resident.LastName = residentFields.LastName
	}
	if residentFields.Phone != "" {
		resident.Phone = residentFields.Phone
	}
	if residentFields.Email != "" {
		resident.Email = residentFields.Email
	}
	if residentFields.Password != "" {
		resident.Password = residentFields.Password
	}
	if residentFields.UnlimDays != nil {
		resident.UnlimDays = residentFields.UnlimDays
	}
	if residentFields.AmtParkingDaysUsed != nil {
		resident.AmtParkingDaysUsed = residentFields.AmtParkingDaysUsed
	}
	if residentFields.TokenVersion != nil {
		resident.TokenVersion = residentFields.TokenVersion
	}
	return nil
}

func (residentRepoMock *ResidentRepoMock) Reset() error {
	residentRepoMock.residents = residentRepoMock.residents[:0]
	return nil
}
