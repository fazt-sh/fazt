import { PageHeader } from '../layout/PageHeader';
import { Card, CardBody } from '../ui';

interface PlaceholderPageProps {
  title: string;
  subtitle?: string;
}

const pageSubtitles = {
  Analytics: "Track visitor stats, page views, and user behavior",
  Redirects: "Manage URL redirects and routing rules",
  Webhooks: "Configure webhook endpoints for events",
  Logs: "View system and application logs",
  Settings: "Configure platform settings and preferences"
};

export function PlaceholderPage({ title, subtitle }: PlaceholderPageProps) {
  const description = subtitle || pageSubtitles[title as keyof typeof pageSubtitles] || "";

  return (
    <div>
      {/* Page Header with title and subtitle */}
      <PageHeader
        title={title}
        description={description}
        action={<div />}
      />

      {/* Grid of empty structural blocks */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main content area */}
        <div className="lg:col-span-2 space-y-6">
          <Card variant="bordered">
            <CardBody className="p-6">
              <div className="h-8 bg-[rgb(var(--bg-subtle))] rounded mb-4 w-1/3"></div>
              <div className="space-y-3">
                <div className="h-4 bg-[rgb(var(--bg-subtle))] rounded w-full"></div>
                <div className="h-4 bg-[rgb(var(--bg-subtle))] rounded w-5/6"></div>
                <div className="h-4 bg-[rgb(var(--bg-subtle))] rounded w-4/6"></div>
              </div>
            </CardBody>
          </Card>

          <Card variant="bordered">
            <CardBody className="p-6">
              <div className="grid grid-cols-2 gap-4">
                <div className="h-32 bg-[rgb(var(--bg-subtle))] rounded"></div>
                <div className="h-32 bg-[rgb(var(--bg-subtle))] rounded"></div>
              </div>
            </CardBody>
          </Card>
        </div>

        {/* Sidebar area */}
        <div className="space-y-6">
          <Card variant="bordered">
            <CardBody className="p-6">
              <div className="h-6 bg-[rgb(var(--bg-subtle))] rounded mb-4 w-2/3"></div>
              <div className="space-y-2">
                <div className="h-3 bg-[rgb(var(--bg-subtle))] rounded w-full"></div>
                <div className="h-3 bg-[rgb(var(--bg-subtle))] rounded w-4/5"></div>
                <div className="h-3 bg-[rgb(var(--bg-subtle))] rounded w-3/5"></div>
              </div>
            </CardBody>
          </Card>

          <Card variant="bordered">
            <CardBody className="p-6">
              <div className="h-20 bg-[rgb(var(--bg-subtle))] rounded"></div>
            </CardBody>
          </Card>
        </div>
      </div>
    </div>
  );
}