import { useState } from 'react';
import { PageHeader } from '../components/layout/PageHeader';
import {
  Button,
  Input,
  Card,
  CardHeader,
  CardBody,
  CardFooter,
  Badge,
  Modal,
  Dropdown,
  DropdownItem,
  DropdownDivider,
  Skeleton,
  Spinner
} from '../components/ui';
import {
  Search,
  User,
  Settings,
  LogOut,
  ChevronDown,
  Mail,
  Lock,
  AlertCircle,
  CheckCircle,
  Info,
  AlertTriangle
} from 'lucide-react';

export function DesignSystemPreline() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [inputValue, setInputValue] = useState('');

  return (
    <div>
      <PageHeader
        title="Preline Design System"
        description="Fazt-branded Preline UI components"
        action={
          <Button onClick={() => setIsModalOpen(true)}>
            Open Demo Modal
          </Button>
        }
      />

      {/* Buttons Section */}
      <Card className="mb-6">
        <CardHeader>
          <h2 className="font-display text-lg text-[rgb(var(--text-primary))]">Buttons</h2>
        </CardHeader>
        <CardBody>
          <div className="space-y-4">
            <div className="flex flex-wrap gap-3">
              <Button>Primary</Button>
              <Button variant="secondary">Secondary</Button>
              <Button variant="ghost">Ghost</Button>
              <Button variant="danger">Danger</Button>
            </div>

            <div className="flex items-center gap-3">
              <span className="text-sm text-[rgb(var(--text-secondary))]">Sizes:</span>
              <Button size="sm">Small</Button>
              <Button size="md">Medium</Button>
              <Button size="lg">Large</Button>
            </div>

            <div className="flex items-center gap-3">
              <Button loading>Loading</Button>
              <Button disabled>Disabled</Button>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Inputs Section */}
      <Card className="mb-6">
        <CardHeader>
          <h2 className="font-display text-lg text-[rgb(var(--text-primary))]">Inputs</h2>
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Default Input"
              placeholder="Enter text..."
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
            />

            <Input
              label="Input with Icon"
              placeholder="Search..."
              icon={<Search className="h-4 w-4" />}
              iconPosition="left"
            />

            <Input
              label="Input with Error"
              error="This field is required"
              icon={<AlertCircle className="h-4 w-4" />}
              iconPosition="right"
            />

            <Input
              label="Disabled Input"
              placeholder="Disabled input"
              disabled
            />
          </div>

          <div className="mt-4">
            <Input
              label="With Helper Text"
              placeholder="Enter email..."
              helperText="We'll never share your email with anyone else."
            />
          </div>
        </CardBody>
      </Card>

      {/* Cards Section */}
      <Card className="mb-6">
        <CardHeader>
          <h2 className="font-display text-lg text-[rgb(var(--text-primary))]">Cards</h2>
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card variant="default" hover>
              <CardBody>
                <h3 className="font-semibold text-[rgb(var(--text-primary))] mb-2">Default Card</h3>
                <p className="text-[rgb(var(--text-secondary))]">This is a default card with hover effect.</p>
              </CardBody>
            </Card>

            <Card variant="bordered" hover>
              <CardBody>
                <h3 className="font-semibold text-[rgb(var(--text-primary))] mb-2">Bordered Card</h3>
                <p className="text-[rgb(var(--text-secondary))]">This card has a visible border.</p>
              </CardBody>
            </Card>

            <Card variant="glass" hover>
              <CardBody>
                <h3 className="font-semibold text-[rgb(var(--text-primary))] mb-2">Glass Card</h3>
                <p className="text-[rgb(var(--text-secondary))]">This card uses glassmorphism.</p>
              </CardBody>
            </Card>
          </div>

          <Card variant="elevated" className="mt-4">
            <CardHeader>
              <h3 className="font-semibold">Card with Sections</h3>
            </CardHeader>
            <CardBody>
              <p>This card has header, body, and footer sections.</p>
            </CardBody>
            <CardFooter>
              <div className="flex justify-end gap-2">
                <Button variant="ghost">Cancel</Button>
                <Button size="sm">Save</Button>
              </div>
            </CardFooter>
          </Card>
        </CardBody>
      </Card>

      {/* Badges Section */}
      <Card className="mb-6">
        <CardHeader>
          <h2 className="font-display text-lg text-[rgb(var(--text-primary))]">Badges</h2>
        </CardHeader>
        <CardBody>
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium text-[rgb(var(--text-secondary))] mb-2">Solid Badges</h4>
              <div className="flex flex-wrap gap-2">
                <Badge>Default</Badge>
                <Badge variant="success">Success</Badge>
                <Badge variant="error">Error</Badge>
                <Badge variant="warning">Warning</Badge>
                <Badge variant="info">Info</Badge>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium text-[rgb(var(--text-secondary))] mb-2">Soft Badges</h4>
              <div className="flex flex-wrap gap-2">
                <Badge soft>Default</Badge>
                <Badge variant="success" soft>Success</Badge>
                <Badge variant="error" soft>Error</Badge>
                <Badge variant="warning" soft>Warning</Badge>
                <Badge variant="info" soft>Info</Badge>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium text-[rgb(var(--text-secondary))] mb-2">With Dots</h4>
              <div className="flex flex-wrap gap-2">
                <Badge dot>Active</Badge>
                <Badge variant="success" dot>Online</Badge>
                <Badge variant="error" dot>Error</Badge>
              </div>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Dropdown Section */}
      <Card className="mb-6">
        <CardHeader>
          <h2 className="font-display text-lg text-[rgb(var(--text-primary))]">Dropdown</h2>
        </CardHeader>
        <CardBody>
          <div className="flex flex-wrap gap-3">
            <Dropdown
              trigger={
                <Button variant="secondary">
                  Dropdown
                  <ChevronDown className="h-4 w-4 ml-2" />
                </Button>
              }
            >
              <DropdownItem icon={<User className="h-4 w-4" />}>Profile</DropdownItem>
              <DropdownItem icon={<Settings className="h-4 w-4" />}>Settings</DropdownItem>
              <DropdownDivider />
              <DropdownItem icon={<LogOut className="h-4 w-4" />}>Logout</DropdownItem>
            </Dropdown>

            <Dropdown
              trigger={
                <Button variant="ghost">
                  Actions
                  <ChevronDown className="h-4 w-4 ml-2" />
                </Button>
              }
              placement="bottom-right"
            >
              <DropdownItem>Edit</DropdownItem>
              <DropdownItem>Duplicate</DropdownItem>
              <DropdownItem disabled>Delete</DropdownItem>
              <DropdownDivider />
              <DropdownItem>Export</DropdownItem>
            </Dropdown>
          </div>
        </CardBody>
      </Card>

      {/* Loading States Section */}
      <Card className="mb-6">
        <CardHeader>
          <h2 className="font-display text-lg text-[rgb(var(--text-primary))]">Loading States</h2>
        </CardHeader>
        <CardBody>
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium text-[rgb(var(--text-secondary))] mb-2">Spinners</h4>
              <div className="flex items-center gap-4">
                <Spinner size="sm" />
                <Spinner size="md" />
                <Spinner size="lg" />
                <Spinner variant="primary" />
                <Spinner variant="success" />
                <Spinner variant="error" />
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium text-[rgb(var(--text-secondary))] mb-2">Skeletons</h4>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p className="text-xs text-[rgb(var(--text-tertiary))] mb-2">Text Skeleton</p>
                  <Skeleton lines={3} />
                </div>
                <div>
                  <p className="text-xs text-[rgb(var(--text-tertiary))] mb-2">Card Skeleton</p>
                  <Skeleton variant="card" />
                </div>
              </div>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Demo Modal"
        size="md"
        footer={
          <>
            <Button variant="ghost" onClick={() => setIsModalOpen(false)}>
              Cancel
            </Button>
            <Button onClick={() => setIsModalOpen(false)}>
              Save Changes
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <p>This is a demo modal showcasing the Preline-based Modal component.</p>
          <Input
            label="Email"
            type="email"
            placeholder="Enter your email"
            icon={<Mail className="h-4 w-4" />}
          />
          <Input
            label="Password"
            type="password"
            placeholder="Enter your password"
            icon={<Lock className="h-4 w-4" />}
          />
        </div>
      </Modal>

      {/* Status Indicators */}
      <Card>
        <CardHeader>
          <h2 className="font-display text-lg text-[rgb(var(--text-primary))]">Status Indicators</h2>
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="text-center p-4 rounded-lg border border-[rgb(var(--border-primary))]">
              <CheckCircle className="h-8 w-8 text-[rgb(var(--success))] mx-auto mb-2" />
              <p className="text-sm font-medium">Success</p>
            </div>
            <div className="text-center p-4 rounded-lg border border-[rgb(var(--border-primary))]">
              <AlertTriangle className="h-8 w-8 text-[rgb(var(--warning))] mx-auto mb-2" />
              <p className="text-sm font-medium">Warning</p>
            </div>
            <div className="text-center p-4 rounded-lg border border-[rgb(var(--border-primary))]">
              <AlertCircle className="h-8 w-8 text-[rgb(var(--error))] mx-auto mb-2" />
              <p className="text-sm font-medium">Error</p>
            </div>
            <div className="text-center p-4 rounded-lg border border-[rgb(var(--border-primary))]">
              <Info className="h-8 w-8 text-[rgb(var(--info))] mx-auto mb-2" />
              <p className="text-sm font-medium">Info</p>
            </div>
          </div>
        </CardBody>
      </Card>
    </div>
  );
}