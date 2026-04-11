import {
  type LazyExoticComponent,
  type ComponentType,
  Suspense,
  lazy,
} from "react";
import { createBrowserRouter, Navigate } from "react-router-dom";
import { ProtectedRoute } from "@/lib/auth";
import { Layout } from "@/components/Layout";
import CircularProgress from "@mui/material/CircularProgress";
import Box from "@mui/material/Box";

// Lazy-loaded route components
const LoginPage = lazy(() => import("@/routes/LoginPage"));
const DashboardPage = lazy(() => import("@/routes/DashboardPage"));
const CatalogPage = lazy(() => import("@/routes/CatalogPage"));
const CatalogDetailPage = lazy(() => import("@/routes/CatalogDetailPage"));
const CatalogFormPage = lazy(() => import("@/routes/CatalogFormPage"));
const InventoryPage = lazy(() => import("@/routes/InventoryPage"));
const GroupBuysPage = lazy(() => import("@/routes/GroupBuysPage"));
const GroupBuyDetailPage = lazy(() => import("@/routes/GroupBuyDetailPage"));
const OrdersPage = lazy(() => import("@/routes/OrdersPage"));
const OrderDetailPage = lazy(() => import("@/routes/OrderDetailPage"));
const ProcurementPage = lazy(() => import("@/routes/ProcurementPage"));
const SuppliersPage = lazy(() => import("@/routes/SuppliersPage"));
const PurchaseOrdersPage = lazy(() => import("@/routes/PurchaseOrdersPage"));
const PurchaseOrderDetailPage = lazy(
  () => import("@/routes/PurchaseOrderDetailPage"),
);
const LandedCostsPage = lazy(() => import("@/routes/LandedCostsPage"));
const ReportsPage = lazy(() => import("@/routes/ReportsPage"));
const AdminPage = lazy(() => import("@/routes/AdminPage"));
const UsersPage = lazy(() => import("@/routes/UsersPage"));
const AuditPage = lazy(() => import("@/routes/AuditPage"));
const BackupsPage = lazy(() => import("@/routes/BackupsPage"));
const BiometricPage = lazy(() => import("@/routes/BiometricPage"));
const VariancesPage = lazy(() => import("@/routes/VariancesPage"));
const LocationsPage = lazy(() => import("@/routes/LocationsPage"));
const MembersPage = lazy(() => import("@/routes/MembersPage"));
const CoachesPage = lazy(() => import("@/routes/CoachesPage"));

type UserRole =
  | "administrator"
  | "operations_manager"
  | "procurement_specialist"
  | "coach"
  | "member";

function LoadingFallback() {
  return (
    <Box
      sx={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        minHeight: "50vh",
      }}
    >
      <CircularProgress />
    </Box>
  );
}

function withSuspense(Component: LazyExoticComponent<ComponentType>) {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <Component />
    </Suspense>
  );
}

function withRole(
  Component: LazyExoticComponent<ComponentType>,
  allowedRoles: UserRole[],
) {
  return (
    <ProtectedRoute allowedRoles={allowedRoles}>
      <Suspense fallback={<LoadingFallback />}>
        <Component />
      </Suspense>
    </ProtectedRoute>
  );
}

// Protected layout wrapper — used as parent for all authenticated routes.
function AuthedLayout() {
  return (
    <ProtectedRoute>
      <Layout />
    </ProtectedRoute>
  );
}

export const router = createBrowserRouter([
  // Public
  {
    path: "/login",
    element: withSuspense(LoginPage),
  },

  // Authenticated shell — all protected routes share the Layout (sidebar + top bar)
  {
    path: "/",
    element: <AuthedLayout />,
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: "dashboard",
        // members do not have dashboard access — match ROLE_PERMISSIONS in constants.ts
        element: withRole(DashboardPage, [
          "administrator",
          "operations_manager",
          "procurement_specialist",
          "coach",
        ]),
      },
      {
        path: "catalog",
        element: withRole(CatalogPage, [
          "administrator",
          "operations_manager",
          "procurement_specialist",
          "member",
        ]),
      },
      {
        path: "catalog/new",
        element: withRole(CatalogFormPage, [
          "administrator",
          "operations_manager",
        ]),
      },
      {
        path: "catalog/:id",
        element: withRole(CatalogDetailPage, [
          "administrator",
          "operations_manager",
          "procurement_specialist",
          "member",
        ]),
      },
      {
        path: "catalog/:id/edit",
        element: withRole(CatalogFormPage, [
          "administrator",
          "operations_manager",
        ]),
      },
      {
        path: "inventory",
        element: withRole(InventoryPage, [
          "administrator",
          "operations_manager",
        ]),
      },
      {
        path: "group-buys",
        element: withSuspense(GroupBuysPage),
      },
      {
        path: "group-buys/:id",
        element: withSuspense(GroupBuyDetailPage),
      },
      {
        path: "orders",
        element: withSuspense(OrdersPage),
      },
      {
        path: "orders/:id",
        element: withSuspense(OrderDetailPage),
      },
      {
        path: "procurement",
        element: withRole(ProcurementPage, [
          "administrator",
          "operations_manager",
          "procurement_specialist",
        ]),
      },
      {
        path: "procurement/suppliers",
        element: withRole(SuppliersPage, [
          "administrator",
          "operations_manager",
          "procurement_specialist",
        ]),
      },
      {
        path: "procurement/purchase-orders",
        element: withRole(PurchaseOrdersPage, [
          "administrator",
          "operations_manager",
          "procurement_specialist",
        ]),
      },
      {
        path: "procurement/purchase-orders/:id",
        element: withRole(PurchaseOrderDetailPage, [
          "administrator",
          "operations_manager",
          "procurement_specialist",
        ]),
      },
      {
        path: "procurement/variances",
        element: withRole(VariancesPage, [
          "administrator",
          "operations_manager",
          "procurement_specialist",
        ]),
      },
      {
        path: "procurement/landed-costs",
        element: withRole(LandedCostsPage, [
          "administrator",
          "operations_manager",
          "procurement_specialist",
        ]),
      },
      {
        path: "reports",
        element: withSuspense(ReportsPage),
      },
      {
        path: "admin",
        element: withRole(AdminPage, ["administrator"]),
      },
      {
        path: "admin/users",
        element: withRole(UsersPage, ["administrator"]),
      },
      {
        path: "admin/audit",
        element: withRole(AuditPage, ["administrator"]),
      },
      {
        path: "admin/backups",
        element: withRole(BackupsPage, ["administrator"]),
      },
      {
        path: "admin/biometric",
        element: withRole(BiometricPage, ["administrator"]),
      },
      {
        path: "locations",
        element: withRole(LocationsPage, [
          "administrator",
          "operations_manager",
        ]),
      },
      {
        path: "members",
        element: withRole(MembersPage, ["administrator", "operations_manager"]),
      },
      {
        path: "coaches",
        element: withRole(CoachesPage, ["administrator", "operations_manager"]),
      },
    ],
  },
]);
