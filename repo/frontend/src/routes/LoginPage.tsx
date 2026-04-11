import { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import { useAuth } from '@/lib/auth';
import { loginSchema } from '@/lib/validation';
import type { ApiError } from '@/lib/api-client';
import type { z } from 'zod';

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const { login, isAuthenticated, isLoading, lockoutState, captchaState, verifyCaptcha } = useAuth();
  const [captchaAnswer, setCaptchaAnswer] = useState('');
  const [captchaError, setCaptchaError] = useState('');
  const [isVerifying, setIsVerifying] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: { pathname: string } })?.from?.pathname ?? '/dashboard';

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  // Redirect if already authenticated
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, isLoading, navigate, from]);

  const onSubmit = async (data: LoginFormData) => {
    try {
      await login(data.email, data.password);
      navigate(from, { replace: true });
    } catch (err) {
      const apiErr = err as ApiError;
      if (apiErr.status === 401) {
        setError('password', { message: 'Invalid email or password' });
      }
      // Lockout and captcha states are handled by AuthContext
    }
  };

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        bgcolor: 'background.default',
        p: 3,
      }}
    >
      <Box sx={{ mb: 3, textAlign: 'center' }}>
        <Typography variant="h4" component="h1" fontWeight={700} color="primary.dark">
          FitCommerce
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Operations &amp; Inventory Suite
        </Typography>
      </Box>

      <Card sx={{ width: '100%', maxWidth: 400 }}>
        <CardContent sx={{ p: 4 }}>
          <Typography variant="h5" fontWeight={600} gutterBottom>
            Sign in
          </Typography>

          {lockoutState && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Your account is locked until{' '}
              {lockoutState.lockedUntil
                ? new Date(lockoutState.lockedUntil).toLocaleTimeString()
                : 'a later time'}
              . Contact an administrator if this is unexpected.
            </Alert>
          )}

          {captchaState?.required && (
            <Box sx={{ mb: 2 }}>
              <Alert severity="warning" sx={{ mb: 1 }}>
                Security verification required. Answer the question below to continue.
              </Alert>
              {captchaState.challengeData && (
                <Typography variant="body2" sx={{ mb: 1, fontWeight: 600 }}>
                  {captchaState.challengeData}
                </Typography>
              )}
              <TextField
                label="Your answer"
                value={captchaAnswer}
                onChange={e => {
                  setCaptchaAnswer(e.target.value);
                  setCaptchaError('');
                }}
                error={Boolean(captchaError)}
                helperText={captchaError || ' '}
                fullWidth
                disabled={isVerifying}
                autoFocus
                sx={{ mb: 1 }}
              />
              <Button
                variant="outlined"
                fullWidth
                disabled={isVerifying || !captchaAnswer.trim()}
                onClick={async () => {
                  setIsVerifying(true);
                  try {
                    await verifyCaptcha(captchaState.challengeId, captchaAnswer.trim());
                    setCaptchaAnswer('');
                    setCaptchaError('');
                  } catch {
                    setCaptchaError('Incorrect answer. Please try again.');
                  } finally {
                    setIsVerifying(false);
                  }
                }}
              >
                {isVerifying ? <CircularProgress size={20} color="inherit" /> : 'Verify'}
              </Button>
            </Box>
          )}

          <Box
            component="form"
            onSubmit={handleSubmit(onSubmit)}
            noValidate
            sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}
          >
            <TextField
              {...register('email')}
              label="Email address"
              type="email"
              autoComplete="email"
              autoFocus
              error={Boolean(errors.email)}
              helperText={errors.email?.message}
              fullWidth
              disabled={isSubmitting || Boolean(lockoutState)}
            />

            <TextField
              {...register('password')}
              label="Password"
              type="password"
              autoComplete="current-password"
              error={Boolean(errors.password)}
              helperText={errors.password?.message}
              fullWidth
              disabled={isSubmitting || Boolean(lockoutState)}
            />

            <Button
              type="submit"
              variant="contained"
              fullWidth
              disabled={isSubmitting || Boolean(lockoutState)}
              sx={{ mt: 1, position: 'relative' }}
            >
              {isSubmitting ? (
                <>
                  <CircularProgress size={20} color="inherit" sx={{ mr: 1 }} />
                  Signing in…
                </>
              ) : (
                'Sign in'
              )}
            </Button>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
}
