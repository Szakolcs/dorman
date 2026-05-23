package administration

import (
	"dorm-man/internal/models"
	"fmt"

	"github.com/google/uuid"
)

// MassAssignmentService runs tenant mass-assignment workflows from the
// administration tenants page.
type MassAssignmentService struct {
	service *Service
}

func NewMassAssignmentService(service *Service) *MassAssignmentService {
	return &MassAssignmentService{service: service}
}

func (s *MassAssignmentService) Run(actorID uuid.UUID, req TenantMassAssignmentRequest) error {
	if actorID == uuid.Nil {
		return ErrUnauthorized
	}

	tenants, err := s.service.ListActiveTenants()
	if err != nil {
		return err
	}
	tenants = filterUnassignedTenants(tenants)
	if len(tenants) == 0 {
		return fmt.Errorf("%w: no unassigned active tenants", ErrValidation)
	}

	flats, err := s.service.ListFlats()
	if err != nil {
		return err
	}
	rooms, err := s.service.ListRooms()
	if err != nil {
		return err
	}

	massTenants, err := tenantsToMassAssignment(tenants, req)
	if err != nil {
		return err
	}
	massFlats := flatsToMassAssignment(flats, rooms)
	if len(massFlats) == 0 {
		return fmt.Errorf("%w: no available rooms", ErrValidation)
	}

	results, err := RunMassAssignment(req, massTenants, massFlats)
	if err != nil {
		return err
	}

	assignments := make([]AssignRoomRequest, 0)
	for _, result := range results {
		roomID, err := uuid.Parse(result.RoomID)
		if err != nil {
			return fmt.Errorf("%w: invalid room id %q", ErrValidation, result.RoomID)
		}
		for _, tenant := range result.Tenants {
			tenantID, err := uuid.Parse(tenant.ID)
			if err != nil {
				return fmt.Errorf("%w: invalid tenant id %q", ErrValidation, tenant.ID)
			}
			assignments = append(assignments, AssignRoomRequest{
				TenantID: tenantID,
				RoomID:   roomID,
			})
		}
	}

	if len(assignments) == 0 {
		return fmt.Errorf("%w: mass assignment produced no room assignments", ErrValidation)
	}

	_, err = s.service.MassAssignRooms(actorID, MassAssignRequest{Assignments: assignments})
	return err
}

func tenantsToMassAssignment(tenants []models.Tenant, req TenantMassAssignmentRequest) ([]MassAssignmentTenant, error) {
	requiredAttrs := append(parseAttributeList(req.StrictGroups), parseAttributeList(req.Preferences)...)
	required := make(map[string]struct{}, len(requiredAttrs))
	for _, attr := range requiredAttrs {
		required[attr] = struct{}{}
	}

	out := make([]MassAssignmentTenant, 0, len(tenants))
	for _, tenant := range tenants {
		massTenant, err := tenantToMassAssignment(tenant, required)
		if err != nil {
			return nil, err
		}
		out = append(out, massTenant)
	}
	return out, nil
}

func tenantToMassAssignment(tenant models.Tenant, required map[string]struct{}) (MassAssignmentTenant, error) {
	massTenant := MassAssignmentTenant{
		ID: tenant.ID.String(),
	}

	if _, ok := required["sex"]; ok {
		if tenant.Sex == nil {
			return MassAssignmentTenant{}, fmt.Errorf("%w: tenant %s is missing sex", ErrValidation, tenant.StudentCode)
		}
		massTenant.Sex = *tenant.Sex
	} else if tenant.Sex != nil {
		massTenant.Sex = *tenant.Sex
	}

	if _, ok := required["nationality"]; ok {
		if tenant.Nationality == nil {
			return MassAssignmentTenant{}, fmt.Errorf("%w: tenant %s is missing nationality", ErrValidation, tenant.StudentCode)
		}
		massTenant.Nationality = *tenant.Nationality
	} else if tenant.Nationality != nil {
		massTenant.Nationality = *tenant.Nationality
	}

	if _, ok := required["degree"]; ok {
		if tenant.Degree == nil {
			return MassAssignmentTenant{}, fmt.Errorf("%w: tenant %s is missing degree", ErrValidation, tenant.StudentCode)
		}
		massTenant.Degree = *tenant.Degree
	} else if tenant.Degree != nil {
		massTenant.Degree = *tenant.Degree
	}

	if _, ok := required["faculty"]; ok {
		if tenant.Faculty == nil {
			return MassAssignmentTenant{}, fmt.Errorf("%w: tenant %s is missing faculty", ErrValidation, tenant.StudentCode)
		}
		massTenant.Faculty = *tenant.Faculty
	} else if tenant.Faculty != nil {
		massTenant.Faculty = *tenant.Faculty
	}

	if _, ok := required["age"]; ok {
		if tenant.Age == nil {
			return MassAssignmentTenant{}, fmt.Errorf("%w: tenant %s is missing age", ErrValidation, tenant.StudentCode)
		}
		massTenant.Age = *tenant.Age
	} else if tenant.Age != nil {
		massTenant.Age = *tenant.Age
	}

	return massTenant, nil
}

func filterUnassignedTenants(tenants []models.Tenant) []models.Tenant {
	out := make([]models.Tenant, 0, len(tenants))
	for _, tenant := range tenants {
		assigned := false
		for _, assignment := range tenant.RoomAssignments {
			if assignment.EndedAt == nil {
				assigned = true
				break
			}
		}
		if !assigned {
			out = append(out, tenant)
		}
	}
	return out
}

func flatsToMassAssignment(flats []models.Flat, rooms []models.Room) []MassAssignmentFlat {
	roomsByFlat := make(map[uuid.UUID][]models.Room, len(flats))
	for _, room := range rooms {
		roomsByFlat[room.FlatID] = append(roomsByFlat[room.FlatID], room)
	}

	out := make([]MassAssignmentFlat, 0, len(flats))
	for _, flat := range flats {
		flatRooms := roomsByFlat[flat.ID]
		if len(flatRooms) == 0 {
			continue
		}
		massFlat := MassAssignmentFlat{
			ID:    flat.ID.String(),
			Rooms: make([]MassAssignmentRoom, 0, len(flatRooms)),
		}
		for _, room := range flatRooms {
			occupied := len(room.Assignments)
			available := room.Capacity - occupied
			if available <= 0 {
				continue
			}
			massFlat.Rooms = append(massFlat.Rooms, MassAssignmentRoom{
				ID:       room.ID.String(),
				Capacity: available,
			})
			massFlat.Capacity += available
		}
		if len(massFlat.Rooms) > 0 {
			out = append(out, massFlat)
		}
	}
	return out
}
