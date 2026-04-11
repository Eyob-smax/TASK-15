import { Component, type ReactNode, type ErrorInfo } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Alert from '@mui/material/Alert';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack);
  }

  reset = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return <>{this.props.fallback}</>;
      }

      return (
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '40vh',
            p: 4,
            gap: 2,
          }}
        >
          <Alert severity="error" sx={{ maxWidth: 560, width: '100%' }}>
            <Typography variant="subtitle1" fontWeight={600} gutterBottom>
              Something went wrong
            </Typography>
            <Typography variant="body2" sx={{ wordBreak: 'break-word' }}>
              {this.state.error?.message ?? 'An unexpected error occurred.'}
            </Typography>
          </Alert>
          <Button variant="outlined" onClick={this.reset}>
            Try again
          </Button>
        </Box>
      );
    }

    return this.props.children;
  }
}
