import { useState } from 'react'
import { CreateBasePlaylistForm } from '../playlists/CreateBasePlaylistForm'
import { Card, CardBody, CardTitle, CardActions, Button } from '../../components/ui'

export function DashboardCards() {
  const [showCreateForm, setShowCreateForm] = useState(false)

  const handleCreateSuccess = () => {
    setShowCreateForm(false)
    // TODO: Add success notification
    alert('Base playlist created successfully!')
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
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <Card>
        <CardBody>
          <CardTitle>Base Playlists</CardTitle>
          <p>Create and manage your source playlists</p>
          <CardActions>
            <Button onClick={() => setShowCreateForm(true)}>
              Create Base Playlist
            </Button>
          </CardActions>
        </CardBody>
      </Card>
      
      <Card>
        <CardBody>
          <CardTitle>Child Playlists</CardTitle>
          <p>Configure filtered playlists (coming soon)</p>
          <CardActions>
            <Button disabled>View</Button>
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
  )
}