import { useState } from 'react';
import { PageHeader } from '../components/layout/PageHeader';
import { Card, CardHeader, CardBody, CardFooter } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Badge } from '../components/ui/Badge';
import { Skeleton, SkeletonText } from '../components/ui/Skeleton';
import { Spinner } from '../components/ui/Spinner';
import { Modal, ModalFooter } from '../components/ui/Modal';
import { Dropdown, DropdownItem, DropdownDivider } from '../components/ui/Dropdown';
import { Home, Settings, LogOut } from 'lucide-react';

export function DesignSystem() {
  const [modalOpen, setModalOpen] = useState(false);

  return (
    <div>
      <PageHeader
        title="Design System"
        description="Component showcase and documentation"
      />

      <div className="space-y-8">
        {/* Colors */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Colors</h2>
          </CardHeader>
          <CardBody>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <div className="h-16 rounded-lg bg-primary mb-2"></div>
                <p className="text-sm font-medium">Primary</p>
                <p className="text-xs text-gray-500">rgb(255, 149, 0)</p>
              </div>
              <div>
                <div className="h-16 rounded-lg bg-green-500 mb-2"></div>
                <p className="text-sm font-medium">Success</p>
                <p className="text-xs text-gray-500">rgb(34, 197, 94)</p>
              </div>
              <div>
                <div className="h-16 rounded-lg bg-red-500 mb-2"></div>
                <p className="text-sm font-medium">Error</p>
                <p className="text-xs text-gray-500">rgb(239, 68, 68)</p>
              </div>
              <div>
                <div className="h-16 rounded-lg bg-blue-500 mb-2"></div>
                <p className="text-sm font-medium">Info</p>
                <p className="text-xs text-gray-500">rgb(59, 130, 246)</p>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Typography */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Typography</h2>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              <div>
                <h1 className="text-4xl font-bold">Heading 1</h1>
                <p className="text-xs text-gray-500">text-4xl font-bold</p>
              </div>
              <div>
                <h2 className="text-3xl font-bold">Heading 2</h2>
                <p className="text-xs text-gray-500">text-3xl font-bold</p>
              </div>
              <div>
                <h3 className="text-2xl font-semibold">Heading 3</h3>
                <p className="text-xs text-gray-500">text-2xl font-semibold</p>
              </div>
              <div>
                <p className="text-base">Body text - The quick brown fox jumps over the lazy dog.</p>
                <p className="text-xs text-gray-500">text-base</p>
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Small text - The quick brown fox jumps over the lazy dog.
                </p>
                <p className="text-xs text-gray-500">text-sm</p>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Buttons */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Buttons</h2>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              <div>
                <p className="text-sm font-medium mb-2">Variants</p>
                <div className="flex flex-wrap gap-2">
                  <Button variant="primary">Primary</Button>
                  <Button variant="secondary">Secondary</Button>
                  <Button variant="ghost">Ghost</Button>
                  <Button variant="danger">Danger</Button>
                </div>
              </div>
              <div>
                <p className="text-sm font-medium mb-2">Sizes</p>
                <div className="flex flex-wrap items-center gap-2">
                  <Button size="sm">Small</Button>
                  <Button size="md">Medium</Button>
                  <Button size="lg">Large</Button>
                </div>
              </div>
              <div>
                <p className="text-sm font-medium mb-2">States</p>
                <div className="flex flex-wrap gap-2">
                  <Button loading>Loading</Button>
                  <Button disabled>Disabled</Button>
                </div>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Inputs */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Inputs</h2>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              <Input label="Default Input" placeholder="Enter text..." />
              <Input label="With Helper Text" placeholder="Enter text..." helperText="This is helper text" />
              <Input label="With Error" placeholder="Enter text..." error="This field is required" />
              <Input label="Disabled" placeholder="Enter text..." disabled />
            </div>
          </CardBody>
        </Card>

        {/* Badges */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Badges</h2>
          </CardHeader>
          <CardBody>
            <div className="flex flex-wrap gap-2">
              <Badge>Default</Badge>
              <Badge variant="success">Success</Badge>
              <Badge variant="error">Error</Badge>
              <Badge variant="warning">Warning</Badge>
              <Badge variant="info">Info</Badge>
            </div>
          </CardBody>
        </Card>

        {/* Cards */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Cards</h2>
          </CardHeader>
          <CardBody>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Card variant="default">
                <CardBody>Default Card</CardBody>
              </Card>
              <Card variant="bordered">
                <CardBody>Bordered Card</CardBody>
              </Card>
              <Card variant="elevated">
                <CardBody>Elevated Card</CardBody>
              </Card>
            </div>
            <div className="mt-4">
              <Card variant="bordered">
                <CardHeader>Card Header</CardHeader>
                <CardBody>Card Body</CardBody>
                <CardFooter>Card Footer</CardFooter>
              </Card>
            </div>
          </CardBody>
        </Card>

        {/* Skeletons */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Skeletons</h2>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              <div>
                <p className="text-sm font-medium mb-2">Text</p>
                <SkeletonText lines={3} />
              </div>
              <div>
                <p className="text-sm font-medium mb-2">Shapes</p>
                <div className="flex gap-4">
                  <Skeleton variant="circle" width="40px" height="40px" />
                  <Skeleton variant="rect" width="200px" height="40px" />
                </div>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Spinners */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Spinners</h2>
          </CardHeader>
          <CardBody>
            <div className="flex items-center gap-4">
              <Spinner size="sm" />
              <Spinner size="md" />
              <Spinner size="lg" />
            </div>
          </CardBody>
        </Card>

        {/* Modal */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Modal</h2>
          </CardHeader>
          <CardBody>
            <Button onClick={() => setModalOpen(true)}>Open Modal</Button>
            <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)} title="Example Modal">
              <p className="text-gray-600 dark:text-gray-400">
                This is an example modal dialog. It uses Headless UI for accessibility and animations.
              </p>
              <ModalFooter>
                <Button variant="ghost" onClick={() => setModalOpen(false)}>
                  Cancel
                </Button>
                <Button variant="primary" onClick={() => setModalOpen(false)}>
                  Confirm
                </Button>
              </ModalFooter>
            </Modal>
          </CardBody>
        </Card>

        {/* Dropdown */}
        <Card variant="bordered">
          <CardHeader>
            <h2 className="text-lg font-semibold">Dropdown</h2>
          </CardHeader>
          <CardBody>
            <Dropdown trigger={<Button>Open Dropdown</Button>}>
              <DropdownItem icon={<Home className="h-4 w-4" />}>
                Home
              </DropdownItem>
              <DropdownItem icon={<Settings className="h-4 w-4" />}>
                Settings
              </DropdownItem>
              <DropdownDivider />
              <DropdownItem icon={<LogOut className="h-4 w-4" />}>
                Logout
              </DropdownItem>
            </Dropdown>
          </CardBody>
        </Card>
      </div>
    </div>
  );
}
