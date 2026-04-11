import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import Snackbar from '@mui/material/Snackbar';
import Alert, { type AlertColor } from '@mui/material/Alert';

interface Notification {
  id: number;
  message: string;
  severity: AlertColor;
}

interface NotifyFn {
  success: (message: string) => void;
  error: (message: string) => void;
  warning: (message: string) => void;
  info: (message: string) => void;
}

const NotifyContext = createContext<NotifyFn | null>(null);

let nextId = 0;

export function NotificationsProvider({ children }: { children: ReactNode }) {
  const [queue, setQueue] = useState<Notification[]>([]);
  const [current, setCurrent] = useState<Notification | null>(null);
  const [open, setOpen] = useState(false);

  const enqueue = useCallback((message: string, severity: AlertColor) => {
    const notification: Notification = { id: ++nextId, message, severity };
    setQueue(prev => {
      if (prev.length === 0 && !open) {
        setCurrent(notification);
        setOpen(true);
        return [];
      }
      return [...prev, notification];
    });
  }, [open]);

  const handleClose = (_: unknown, reason?: string) => {
    if (reason === 'clickaway') return;
    setOpen(false);
  };

  const handleExited = () => {
    if (queue.length > 0) {
      const [next, ...rest] = queue;
      setQueue(rest);
      setCurrent(next);
      setOpen(true);
    } else {
      setCurrent(null);
    }
  };

  const notify: NotifyFn = {
    success: msg => enqueue(msg, 'success'),
    error: msg => enqueue(msg, 'error'),
    warning: msg => enqueue(msg, 'warning'),
    info: msg => enqueue(msg, 'info'),
  };

  return (
    <NotifyContext.Provider value={notify}>
      {children}
      <Snackbar
        open={open}
        autoHideDuration={5000}
        onClose={handleClose}
        TransitionProps={{ onExited: handleExited }}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={() => setOpen(false)}
          severity={current?.severity ?? 'info'}
          variant="filled"
          sx={{ width: '100%', minWidth: 280 }}
        >
          {current?.message}
        </Alert>
      </Snackbar>
    </NotifyContext.Provider>
  );
}

export function useNotify(): NotifyFn {
  const ctx = useContext(NotifyContext);
  if (!ctx) throw new Error('useNotify must be used within NotificationsProvider');
  return ctx;
}
