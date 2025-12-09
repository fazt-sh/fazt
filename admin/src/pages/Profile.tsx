import { useState } from 'react';
import { PageHeader } from '../components/layout/PageHeader';
import { Breadcrumbs } from '../components/ui/Breadcrumbs';
import { Button, Card, CardBody, Input, Tabs, SectionHeader } from '../components/ui';
import { User, Camera, Mail, AtSign } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import { useForm } from 'react-hook-form';

export function Profile() {
  const { user } = useAuth();
  const { success } = useToast();
  
  // Mock state for user details (since auth context usually just has basic info)
  const [profile, setProfile] = useState({
    displayName: 'Admin User',
    email: 'admin@fazt.sh',
    photoUrl: null
  });

  const GeneralTab = () => {
    const { register, handleSubmit } = useForm({
      defaultValues: profile
    });

    const onSubmit = (data: any) => {
      setProfile({ ...profile, ...data });
      success('Profile updated successfully');
    };

    return (
      <div className="max-w-2xl">
        <Card variant="bordered" className="mb-6">
          <CardBody>
            <div className="flex items-center gap-6 mb-8">
              <div className="relative group">
                <div className="w-20 h-20 rounded-full bg-gradient-to-br from-[rgb(var(--accent-start))] to-[rgb(var(--accent-mid))] flex items-center justify-center text-white text-3xl font-bold shadow-lg">
                  {user?.username?.[0]?.toUpperCase() || 'A'}
                </div>
                <button className="absolute bottom-0 right-0 p-1.5 bg-[rgb(var(--bg-elevated))] border border-[rgb(var(--border-primary))] rounded-full shadow-sm hover:bg-[rgb(var(--bg-hover))] transition-colors text-[rgb(var(--text-secondary))]">
                  <Camera className="w-4 h-4" />
                </button>
              </div>
              <SectionHeader
                title="Profile Photo"
                description="This will be displayed on your profile."
                className="mb-0"
              />
            </div>

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
                    Display Name
                  </label>
                  <div className="relative">
                    <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[rgb(var(--text-tertiary))]" />
                    <Input
                      {...register('displayName')}
                      className="pl-9"
                      placeholder="John Doe"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
                    Username
                  </label>
                  <div className="relative">
                    <AtSign className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[rgb(var(--text-tertiary))]" />
                    <Input
                      value={user?.username}
                      disabled
                      className="pl-9 bg-[rgb(var(--bg-subtle))]"
                    />
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
                  Email Address
                </label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[rgb(var(--text-tertiary))]" />
                  <Input
                    {...register('email')}
                    type="email"
                    className="pl-9"
                    placeholder="john@example.com"
                  />
                </div>
              </div>

              <div className="flex justify-end pt-4">
                <Button type="submit" variant="primary">
                  Save Changes
                </Button>
              </div>
            </form>
          </CardBody>
        </Card>
      </div>
    );
  };

  const SecurityTab = () => {
    const { register, handleSubmit, reset } = useForm();

    const onSubmit = () => {
      success('Password updated successfully');
      reset();
    };

    return (
      <div className="max-w-2xl">
        <Card variant="bordered">
          <CardBody>
            <SectionHeader
              title="Change Password"
              description="Ensure your account is using a long, random password to stay secure."
            />

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
                  Current Password
                </label>
                <Input
                  {...register('currentPassword')}
                  type="password"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
                  New Password
                </label>
                <Input
                  {...register('newPassword')}
                  type="password"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
                  Confirm New Password
                </label>
                <Input
                  {...register('confirmPassword')}
                  type="password"
                />
              </div>

              <div className="flex justify-end pt-4">
                <Button type="submit" variant="primary">
                  Update Password
                </Button>
              </div>
            </form>
          </CardBody>
        </Card>
      </div>
    );
  };

  return (
    <div className="animate-fade-in">
      <Breadcrumbs />
      <PageHeader
        title="Profile Settings"
        description="Manage your account settings and preferences."
      />

      <Tabs
        defaultTab="general"
        tabs={[
          { id: 'general', label: 'General', content: <GeneralTab /> },
          { id: 'security', label: 'Security', content: <SecurityTab /> },
        ]}
      />
    </div>
  );
}
