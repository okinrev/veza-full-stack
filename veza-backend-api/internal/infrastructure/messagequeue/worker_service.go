package messagequeue

// BackgroundWorkerService pour les tâches en arrière-plan
type BackgroundWorkerService struct {
	natsService *NATSService
}

// NewBackgroundWorkerService crée un nouveau service worker
func NewBackgroundWorkerService() *BackgroundWorkerService {
	return &BackgroundWorkerService{}
}
