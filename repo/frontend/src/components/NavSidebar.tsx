import { useCallback, type ReactNode } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import Box from '@mui/material/Box';
import Drawer from '@mui/material/Drawer';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Divider from '@mui/material/Divider';
import Typography from '@mui/material/Typography';

import DashboardIcon from '@mui/icons-material/Dashboard';
import Inventory2Icon from '@mui/icons-material/Inventory2';
import WarehouseIcon from '@mui/icons-material/Warehouse';
import GroupsIcon from '@mui/icons-material/Groups';
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart';
import BusinessIcon from '@mui/icons-material/Business';
import AssessmentIcon from '@mui/icons-material/Assessment';
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings';

import { useAuth } from '@/lib/auth';
import { ROLE_PERMISSIONS } from '@/lib/constants';

export const DRAWER_WIDTH = 240;

interface NavItem {
  key: string;
  label: string;
  icon: ReactNode;
  path: string;
  matchPaths?: string[];
}

const ALL_NAV_ITEMS: NavItem[] = [
  {
    key: 'dashboard',
    label: 'Dashboard',
    icon: <DashboardIcon />,
    path: '/dashboard',
  },
  {
    key: 'catalog',
    label: 'Catalog',
    icon: <Inventory2Icon />,
    path: '/catalog',
  },
  {
    key: 'inventory',
    label: 'Inventory',
    icon: <WarehouseIcon />,
    path: '/inventory',
  },
  {
    key: 'group_buys',
    label: 'Group Buys',
    icon: <GroupsIcon />,
    path: '/group-buys',
  },
  {
    key: 'orders',
    label: 'Orders',
    icon: <ShoppingCartIcon />,
    path: '/orders',
  },
  {
    key: 'procurement',
    label: 'Procurement',
    icon: <BusinessIcon />,
    path: '/procurement',
  },
  {
    key: 'reports',
    label: 'Reports',
    icon: <AssessmentIcon />,
    path: '/reports',
  },
  {
    key: 'admin',
    label: 'Admin',
    icon: <AdminPanelSettingsIcon />,
    path: '/admin',
    matchPaths: ['/admin', '/admin/users', '/admin/audit', '/admin/backups', '/admin/biometric'],
  },
];

function isActive(item: NavItem, pathname: string): boolean {
  if (item.matchPaths) {
    return item.matchPaths.some(p => pathname === p || pathname.startsWith(p + '/'));
  }
  return pathname === item.path || pathname.startsWith(item.path + '/');
}

export function NavSidebar() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const allowedModules = user ? (ROLE_PERMISSIONS[user.role] ?? []) : [];

  const visibleItems = ALL_NAV_ITEMS.filter(item =>
    allowedModules.includes(item.key) ||
    (item.key === 'admin' && allowedModules.some(m => ['admin', 'audit', 'backups', 'biometric', 'users'].includes(m))),
  );

  const handleNav = useCallback(
    (path: string) => {
      navigate(path);
    },
    [navigate],
  );

  return (
    <Drawer
      variant="permanent"
      sx={{
        width: DRAWER_WIDTH,
        flexShrink: 0,
        '& .MuiDrawer-paper': {
          width: DRAWER_WIDTH,
          boxSizing: 'border-box',
          bgcolor: 'primary.dark',
          color: 'primary.contrastText',
          borderRight: 'none',
        },
      }}
    >
      {/* Brand */}
      <Box
        sx={{
          px: 2,
          py: 2.5,
          borderBottom: '1px solid rgba(255,255,255,0.12)',
        }}
      >
        <Typography variant="h6" fontWeight={700} color="inherit" noWrap>
          FitCommerce
        </Typography>
        <Typography variant="caption" sx={{ opacity: 0.7 }}>
          Operations Suite
        </Typography>
      </Box>

      {/* Nav items */}
      <List sx={{ pt: 1, flex: 1 }}>
        {visibleItems.map(item => {
          const active = isActive(item, location.pathname);
          return (
            <ListItem key={item.key} disablePadding>
                <ListItemButton
                  onClick={() => handleNav(item.path)}
                  selected={active}
                  sx={{
                    mx: 1,
                    my: 0.25,
                    borderRadius: 1,
                    color: active ? 'primary.contrastText' : 'rgba(255,255,255,0.7)',
                    '&.Mui-selected': {
                      bgcolor: 'rgba(255,255,255,0.15)',
                      color: 'primary.contrastText',
                    },
                    '&:hover': {
                      bgcolor: 'rgba(255,255,255,0.1)',
                      color: 'primary.contrastText',
                    },
                    '&.Mui-selected:hover': {
                      bgcolor: 'rgba(255,255,255,0.2)',
                    },
                  }}
                >
                  <ListItemIcon
                    sx={{
                      color: 'inherit',
                      minWidth: 36,
                    }}
                  >
                    {item.icon}
                  </ListItemIcon>
                  <ListItemText
                    primary={item.label}
                    primaryTypographyProps={{ variant: 'body2', fontWeight: active ? 600 : 400 }}
                  />
                </ListItemButton>
            </ListItem>
          );
        })}
      </List>

      <Divider sx={{ borderColor: 'rgba(255,255,255,0.12)' }} />
    </Drawer>
  );
}
