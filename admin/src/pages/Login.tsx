import { useState } from 'react';
import type { FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Input } from '../components/ui/Input';
import { Button } from '../components/ui/Button';
import { Card, CardBody } from '../components/ui/Card';

export function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await login(username, password);
      navigate('/');
    } catch (err: any) {
      setError(err.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[rgb(var(--bg-base))] px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="font-display text-4xl text-[rgb(var(--text-primary))] tracking-tight mb-3">
            Fazt<span className="text-[rgb(var(--accent))]">.sh</span>
          </h1>
          <p className="text-[13px] text-[rgb(var(--text-secondary))]">
            Sign in to your platform
          </p>
        </div>

        <Card variant="bordered" className="overflow-hidden">
          <CardBody className="p-8">
            <form onSubmit={handleSubmit} className="space-y-5">
              <Input
                label="Username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="admin"
                required
                autoFocus
              />

              <Input
                label="Password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
              />

              {error && (
                <div className="p-3 rounded-lg bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-900/50 text-red-600 dark:text-red-400 text-[13px]">
                  {error}
                </div>
              )}

              <Button
                type="submit"
                variant="primary"
                className="w-full mt-6"
                size="lg"
                loading={loading}
              >
                Sign In
              </Button>
            </form>
          </CardBody>
        </Card>

        <p className="text-center text-xs text-[rgb(var(--text-tertiary))] mt-6">
          Personal PaaS • Deploy anywhere
        </p>
      </div>
    </div>
  );
}
