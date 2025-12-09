import { useState } from 'react';
import { PageHeader } from '../../components/layout/PageHeader';
import { Button, Card, CardBody, Badge, Modal, Input } from '../../components/ui';
import { Webhook, Plus, Trash2, Edit2 } from 'lucide-react';
import { useMockMode } from '../../context/MockContext';
import { mockData } from '../../lib/mockData';
import { useForm } from 'react-hook-form';
import type { Webhook as WebhookModel } from '../../types/models';

export function Webhooks() {
  const { enabled: mockMode } = useMockMode();
  const [webhooks, setWebhooks] = useState<WebhookModel[]>(mockMode ? mockData.webhooks : []);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingWebhook, setEditingWebhook] = useState<WebhookModel | null>(null);

  const { register, handleSubmit, reset, setValue } = useForm<{ endpoint: string; method: string }>();

  const openCreateModal = () => {
    setEditingWebhook(null);
    reset({ endpoint: '', method: 'POST' });
    setIsModalOpen(true);
  };

  const openEditModal = (webhook: WebhookModel) => {
    setEditingWebhook(webhook);
    setValue('endpoint', webhook.endpoint);
    setValue('method', webhook.method);
    setIsModalOpen(true);
  };

  const onSubmit = (data: { endpoint: string; method: string }) => {
    if (editingWebhook) {
      setWebhooks(prev => prev.map(w => w.id === editingWebhook.id ? { ...w, ...data } : w));
    } else {
      const newWebhook: WebhookModel = {
        id: `wh_${Date.now()}`,
        endpoint: data.endpoint,
        method: data.method,
        created_at: new Date().toISOString(),
      };
      setWebhooks(prev => [...prev, newWebhook]);
    }
    setIsModalOpen(false);
  };

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this webhook?')) {
      setWebhooks(prev => prev.filter(w => w.id !== id));
    }
  };

  return (
    <div className="animate-fade-in">
      <PageHeader
        title="Webhooks"
        description="Manage incoming and outgoing webhooks."
        action={
          <Button variant="primary" onClick={openCreateModal}>
            <Plus className="w-4 h-4 mr-2" />
            Create Webhook
          </Button>
        }
      />

      <div className="grid gap-4">
        {webhooks.length === 0 ? (
           <Card variant="bordered">
             <CardBody className="text-center py-12">
               <Webhook className="w-12 h-12 text-[rgb(var(--text-tertiary))] mx-auto mb-4" />
               <h3 className="text-lg font-medium text-[rgb(var(--text-primary))]">No webhooks found</h3>
               <p className="text-[rgb(var(--text-secondary))] mt-1 mb-6">Create your first webhook to get started.</p>
               <Button variant="primary" onClick={openCreateModal}>
                 <Plus className="w-4 h-4 mr-2" />
                 Create Webhook
               </Button>
             </CardBody>
           </Card>
        ) : (
          webhooks.map((webhook) => (
            <Card key={webhook.id} variant="bordered" className="hover:border-[rgb(var(--border-secondary))] transition-colors">
              <CardBody className="flex items-center justify-between p-4">
                <div className="flex items-center gap-4">
                  <div className="p-2 bg-[rgb(var(--bg-subtle))] rounded-lg">
                    <Webhook className="w-5 h-5 text-[rgb(var(--accent))]" />
                  </div>
                  <div>
                    <div className="font-mono text-sm text-[rgb(var(--text-primary))]">{webhook.endpoint}</div>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge variant="default" size="sm">{webhook.method}</Badge>
                      <span className="text-xs text-[rgb(var(--text-tertiary))]">
                        Created {new Date(webhook.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button variant="ghost" size="sm" onClick={() => openEditModal(webhook)}>
                    <Edit2 className="w-4 h-4" />
                  </Button>
                  <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20" onClick={() => handleDelete(webhook.id)}>
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
        title={editingWebhook ? 'Edit Webhook' : 'Create Webhook'}
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
              Endpoint URL
            </label>
            <Input
              {...register('endpoint', { required: true })}
              placeholder="https://api.example.com/webhook"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-[rgb(var(--text-secondary))] mb-1">
              Method
            </label>
            <select
              {...register('method')}
              className="w-full px-3 py-2 bg-[rgb(var(--bg-surface))] border border-[rgb(var(--border-primary))] rounded-lg focus:outline-none focus:ring-2 focus:ring-[rgb(var(--accent))] focus:border-transparent text-[rgb(var(--text-primary))]"
            >
              <option value="POST">POST</option>
              <option value="GET">GET</option>
              <option value="PUT">PUT</option>
              <option value="DELETE">DELETE</option>
            </select>
          </div>
          <div className="flex justify-end gap-3 mt-6">
            <Button type="button" variant="ghost" onClick={() => setIsModalOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="primary">
              {editingWebhook ? 'Save Changes' : 'Create Webhook'}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}