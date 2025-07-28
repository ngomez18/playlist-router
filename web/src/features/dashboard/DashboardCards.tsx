import { Card, CardBody, CardTitle, CardActions, Button } from '../../components/ui'

export function DashboardCards() {
  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <Card>
        <CardBody>
          <CardTitle>Base Playlists</CardTitle>
          <p>Manage your source playlists</p>
          <CardActions>
            <Button>View</Button>
          </CardActions>
        </CardBody>
      </Card>
      
      <Card>
        <CardBody>
          <CardTitle>Child Playlists</CardTitle>
          <p>Configure filtered playlists</p>
          <CardActions>
            <Button>View</Button>
          </CardActions>
        </CardBody>
      </Card>
      
      <Card>
        <CardBody>
          <CardTitle>Sync History</CardTitle>
          <p>View recent sync operations</p>
          <CardActions>
            <Button>View</Button>
          </CardActions>
        </CardBody>
      </Card>
    </div>
  )
}