package administration

// MassAssignmentService runs tenant mass-assignment workflows from the
// administration tenants page. Implementation is intentionally empty for now.
type MassAssignmentService struct{}

func NewMassAssignmentService() *MassAssignmentService {
	return &MassAssignmentService{}
}

func (s *MassAssignmentService) Run(req TenantMassAssignmentRequest) error {
	return nil
}
