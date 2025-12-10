import { useState } from 'react';
import { PageHeader } from '../../components/layout/PageHeader';
import { Button, Card, CardBody, Modal, Input } from '../../components/ui';
import { ExternalLink, Plus, Trash2, Edit2, Copy, BarChart3 } from 'lucide-react';
import { useMockMode } from '../../context/MockContext';
import { useToast } from '../../context/ToastContext';
import { mockData } from '../../lib/mockData';
import { useForm } from 'react-hook-form';
import type { Redirect } from '../../types/models';

export function Redirects() {
  const { enabled: mockMode } = useMockMode();
  const { success } = useToast();
  const [redirects, setRedirects] = useState<Redirect[]>(mockMode ? mockData.redirects : []);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingRedirect, setEditingRedirect] = useState<Redirect | null>(null);

  const { register, handleSubmit, reset, setValue } = useForm<{ short_code: string; target_url: string }>();

  const openCreateModal = () => {
    setEditingRedirect(null);
    reset({ short_code: '', target_url: '' });
    setIsModalOpen(true);
  };

  const openEditModal = (redirect: Redirect) => {
    setEditingRedirect(redirect);
    setValue('short_code', redirect.short_code);
    setValue('target_url', redirect.target_url);
    setIsModalOpen(true);
  };

  const onSubmit = (data: { short_code: string; target_url: string }) => {
    if (editingRedirect) {
      setRedirects(prev => prev.map(r => r.id === editingRedirect.id ? { ...r, ...data } : r));
      success('Redirect updated successfully');
    } else {
      const newRedirect: Redirect = {
        id: `r_${Date.now()}`,
        short_code: data.short_code,
        target_url: data.target_url,
        click_count: 0,
        created_at: new Date().toISOString(),
      };
      setRedirects(prev => [...prev, newRedirect]);
      success('Redirect created successfully');
    }
    setIsModalOpen(false);
  };

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this redirect?')) {
      setRedirects(prev => prev.filter(r => r.id !== id));
      success('Redirect deleted');
    }
  };

  const copyLink = (code: string) => {
    const url = `${window.location.origin}/r/${code}`; // Assuming /r/ prefix for redirects
    navigator.clipboard.writeText(url);
    success('Link copied to clipboard');
  };

  return (
    <div className="animate-fade-in">
      <PageHeader
        title="URL Redirects"
        description="Manage short links and URL redirects."
        action={
          <Button variant="primary" onClick={openCreateModal}>
            <Plus className="w-4 h-4 mr-2" />
            Create Redirect
          </Button>
        }
      />

      <div className="grid gap-4">
        {redirects.length === 0 ? (
           <Card variant="bordered">
             <CardBody className="text-center py-12">
               <ExternalLink className="w-12 h-12 text-[rgb(var(--text-tertiary))] mx-auto mb-4" />
               <h3 className="text-lg font-medium text-[rgb(var(--text-primary))]">No redirects found</h3>
               <p className="text-[rgb(var(--text-secondary))] mt-1 mb-6">Create your first redirect to get started.</p>
               <Button variant="primary" onClick={openCreateModal}>
                 <Plus className="w-4 h-4 mr-2" />
                 Create Redirect
               </Button>
             </CardBody>
           </Card>
        ) : (
          redirects.map((redirect) => (
            <Card key={redirect.id} variant="bordered" className="hover:border-[rgb(var(--border-secondary))] transition-colors">
              <CardBody className="flex items-center justify-between p-4">
                <div className="flex items-center gap-4">
                  <div className="p-2 bg-[rgb(var(--bg-subtle))] rounded-lg">
                    <ExternalLink className="w-5 h-5 text-[rgb(var(--accent))]" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                        <span className="font-mono text-sm font-semibold text-[rgb(var(--text-primary))]">/{redirect.short_code}</span>
                        <span className="text-[rgb(var(--text-tertiary))]">→</span>
                        <span className="text-sm text-[rgb(var(--text-secondary))] truncate max-w-md">{redirect.target_url}</span>
                    </div>
                    <div className="flex items-center gap-4 mt-1">
                      <div className="flex items-center gap-1 text-xs text-[rgb(var(--text-tertiary))]">
                        <BarChart3 className="w-3 h-3" />
                        {redirect.click_count} clicks
                      </div>
                      <span className="text-xs text-[rgb(var(--text-tertiary))]">
                        Created {new Date(redirect.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button variant="ghost" size="sm" onClick={() => copyLink(redirect.short_code)}>
                    <Copy className="w-4 h-4" />
                  </Button>
                  <Button variant="ghost" size="sm" onClick={() => openEditModal(redirect)}>
                    <Edit2 className="w-4 h-4" />
                  </Button>
                  <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20" onClick={() => handleDelete(redirect.id)}>
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </CardBody>
            </Card>
          ))
        )}
      </div>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={editingRedirect ? 'Edit Redirect' : 'Create Redirect'}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
              Short Code
            </label>
            <div className="flex items-center">
                <span className="inline-flex items-center px-3 py-2 rounded-l-lg border border-r-0 border-[rgb(var(--border-primary))] bg-[rgb(var(--bg-subtle))] text-[rgb(var(--text-secondary))] text-sm">
                    /
                </span>
                <Input
                {...register('short_code', { required: true })}
                placeholder="twitter"
                className="rounded-l-none"
                />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
              Target URL
            </label>
            <Input
              {...register('target_url', { required: true })}
              placeholder="https://twitter.com/fazt_sh"
            />
          </div>
          <div className="flex justify-end gap-3 mt-6">
            <Button type="button" variant="ghost" onClick={() => setIsModalOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="primary">
              {editingRedirect ? 'Save Changes' : 'Create Redirect'}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}