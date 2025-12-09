import { Card, CardBody } from '../ui';
import { Skeleton } from '../ui';
import { Globe } from 'lucide-react';

export function SitesSkeleton() {
  return (
    <div>
      {/* Page Header skeleton */}
      <div className="flex items-start justify-between mb-8">
        <div>
          <Skeleton variant="rect" width={120} height={40} className="mb-2" />
          <Skeleton variant="text" width={250} height={20} />
        </div>
        <Skeleton variant="rect" width={140} height={40} className="rounded-lg" />
      </div>

      {/* Site cards skeleton grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <Card key={i} variant="bordered" className="hover:shadow-md transition-shadow">
            <CardBody>
              {/* Site header skeleton */}
              <div className="flex items-start justify-between mb-4">
                <div>
                  <Skeleton variant="rect" width={150} height={24} className="mb-1" />
                  <div className="flex items-center gap-2">
                    <Globe className="h-4 w-4 text-gray-400" />
                    <Skeleton variant="text" width={120} height={16} />
                  </div>
                </div>
                <Skeleton variant="rect" width={60} height={24} className="rounded-full" />
              </div>

              {/* Site stats skeleton */}
              <div className="grid grid-cols-2 gap-4 mb-4">
                <div className="flex items-center gap-2">
                  <Skeleton variant="circle" width={16} height={16} />
                  <Skeleton variant="text" width={60} height={16} />
                </div>
                <div className="flex items-center gap-2">
                  <Skeleton variant="circle" width={16} height={16} />
                  <Skeleton variant="text" width={60} height={16} />
                </div>
              </div>

              {/* Last updated skeleton */}
              <Skeleton variant="text" width={100} height={16} className="mb-4" />

              {/* Action buttons skeleton */}
              <div className="flex gap-2">
                <Skeleton variant="rect" height={36} className="flex-1 rounded-lg" />
                <Skeleton variant="rect" width={80} height={36} className="rounded-lg" />
                <Skeleton variant="rect" width={80} height={36} className="rounded-lg" />
              </div>
            </CardBody>
          </Card>
        ))}
      </div>
    </div>
  );
}