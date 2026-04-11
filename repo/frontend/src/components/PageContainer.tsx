import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Breadcrumbs from '@mui/material/Breadcrumbs';
import Link from '@mui/material/Link';
import { Link as RouterLink } from 'react-router-dom';
import type { ReactNode } from 'react';

interface Crumb {
  label: string;
  to?: string;
}

interface PageContainerProps {
  title: string;
  breadcrumbs?: Crumb[];
  actions?: ReactNode;
  children: ReactNode;
}

export function PageContainer({ title, breadcrumbs, actions, children }: PageContainerProps) {
  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <Box
        sx={{
          px: 3,
          py: 2,
          borderBottom: 1,
          borderColor: 'divider',
          bgcolor: 'background.paper',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexShrink: 0,
        }}
      >
        <Box>
          {breadcrumbs && breadcrumbs.length > 0 && (
            <Breadcrumbs sx={{ mb: 0.5 }} aria-label="breadcrumb">
              {breadcrumbs.map((crumb, i) =>
                crumb.to ? (
                  <Link
                    key={i}
                    component={RouterLink}
                    to={crumb.to}
                    underline="hover"
                    color="inherit"
                    variant="caption"
                  >
                    {crumb.label}
                  </Link>
                ) : (
                  <Typography key={i} variant="caption" color="text.primary">
                    {crumb.label}
                  </Typography>
                ),
              )}
            </Breadcrumbs>
          )}
          <Typography variant="h5" component="h1">
            {title}
          </Typography>
        </Box>
        {actions && <Box sx={{ display: 'flex', gap: 1 }}>{actions}</Box>}
      </Box>

      {/* Content */}
      <Box sx={{ flex: 1, overflow: 'auto', p: 3 }}>
        {children}
      </Box>
    </Box>
  );
}
