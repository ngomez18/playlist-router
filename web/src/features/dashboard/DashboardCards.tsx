import { useState } from 'react'
import { CreateBasePlaylistForm } from '../playlists/CreateBasePlaylistForm'
import { Card, CardBody, CardTitle, CardActions, Button, Alert } from '../../components/ui'

interface DashboardCardsProps {
  onNavigateToBasePlaylists: () => void
}

export function DashboardCards({ onNavigateToBasePlaylists }: DashboardCardsProps) {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const handleCreateSuccess = () => {
    setShowCreateForm(false)
    setSuccessMessage('Base playlist created successfully!')
    // Auto-hide success message after 5 seconds
    setTimeout(() => setSuccessMessage(null), 5000)
  }

  if (showCreateForm) {
    return (
      <div className="max-w-md">
        <CreateBasePlaylistForm 
          onSuccess={handleCreateSuccess}
          onCancel={() => setShowCreateForm(false)}
        />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {successMessage && (
        <Alert type="success">
          {successMessage}
        </Alert>
      )}
      
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <Card>
        <CardBody>
          <CardTitle>Base Playlists</CardTitle>
          <p>Set up source playlists to automatically route songs into your themed playlists</p>
          <CardActions>
            <Button onClick={onNavigateToBasePlaylists}>
              Manage
            </Button>
          </CardActions>
        </CardBody>
      </Card>

      <Card>
        <CardBody>
          <CardTitle>Sync History</CardTitle>
          <p>View recent sync operations (coming soon)</p>
          <CardActions>
            <Button disabled>View</Button>
          </CardActions>
        </CardBody>
      </Card>
      </div>
    </div>
  )
}