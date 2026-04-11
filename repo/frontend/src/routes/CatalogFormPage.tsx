import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm, Controller, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Divider from "@mui/material/Divider";
import FormControl from "@mui/material/FormControl";
import FormHelperText from "@mui/material/FormHelperText";
import Grid from "@mui/material/Grid";
import IconButton from "@mui/material/IconButton";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Paper from "@mui/material/Paper";
import Select from "@mui/material/Select";
import Skeleton from "@mui/material/Skeleton";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { OfflineDataNotice } from "@/components/OfflineDataNotice";
import { PageContainer } from "@/components/PageContainer";
import { OFFLINE_MUTATION_MESSAGE, useOfflineStatus } from "@/lib/offline";
import {
  EMPTY_CATALOG_WINDOW,
  getDefaultCatalogFormValues,
  mapCatalogFormToItemPayload,
  mapItemToCatalogFormValues,
} from "@/lib/catalog-form";
import { useItem, useCreateItem, useUpdateItem } from "@/lib/hooks/useItems";
import { useNotify } from "@/lib/notifications";
import { createItemSchema, type CreateItemFormData } from "@/lib/validation";

export default function CatalogFormPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const notify = useNotify();
  const { isOffline } = useOfflineStatus();
  const isEdit = Boolean(id);

  const {
    data: existingItem,
    isLoading: itemLoading,
    error,
    dataUpdatedAt,
  } = useItem(isEdit ? id : undefined);
  const createMutation = useCreateItem();
  const updateMutation = useUpdateItem();

  const {
    register,
    handleSubmit,
    control,
    reset,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<CreateItemFormData>({
    resolver: zodResolver(createItemSchema),
    defaultValues: getDefaultCatalogFormValues(),
  });

  const availabilityWindows = useFieldArray({
    control,
    name: "availability_windows",
  });

  const blackoutWindows = useFieldArray({
    control,
    name: "blackout_windows",
  });

  useEffect(() => {
    if (existingItem) {
      reset(mapItemToCatalogFormValues(existingItem));
    } else if (!isEdit) {
      reset(getDefaultCatalogFormValues());
    }
  }, [existingItem, isEdit, reset]);

  const onSubmit = async (data: CreateItemFormData) => {
    try {
      const payload = mapCatalogFormToItemPayload(data);

      if (isEdit && id) {
        if (!existingItem) {
          throw new Error("Item data is missing");
        }

        await updateMutation.mutateAsync({
          id,
          body: { ...payload, version: existingItem.version },
        });
        notify.success(
          isOffline
            ? "Item update queued. It will sync when you reconnect."
            : "Item updated successfully.",
        );
        navigate(`/catalog/${id}`);
      } else {
        const createdItem = await createMutation.mutateAsync(payload);
        notify.success(
          isOffline
            ? "Item draft queued. It will sync when you reconnect."
            : "Item saved as a draft.",
        );
        if (isOffline) {
          navigate("/catalog");
        } else {
          navigate(createdItem?.id ? `/catalog/${createdItem.id}` : "/catalog");
        }
      }
    } catch (err: unknown) {
      const apiErr = err as { status?: number; message?: string };
      if (apiErr.status === 409) {
        setError("root", {
          message:
            "Item was modified by another user. Please reload and try again.",
        });
      } else if (apiErr.status === 422) {
        setError("root", {
          message: `Validation error: ${apiErr.message ?? "Please check your inputs."}`,
        });
      } else {
        setError("root", { message: "Failed to save item. Please try again." });
      }
    }
  };

  if (isEdit && itemLoading) {
    return (
      <PageContainer
        title="Edit Item"
        breadcrumbs={[{ label: "Catalog", to: "/catalog" }, { label: "Edit" }]}
      >
        <Skeleton variant="rectangular" height={400} />
      </PageContainer>
    );
  }

  const title = isEdit ? `Edit ${existingItem?.name ?? "Item"}` : "Create Item";

  return (
    <PageContainer
      title={title}
      breadcrumbs={[
        { label: "Catalog", to: "/catalog" },
        ...(isEdit && id
          ? [{ label: existingItem?.name ?? id, to: `/catalog/${id}` }]
          : []),
        { label: isEdit ? "Edit" : "Create" },
      ]}
    >
      <Paper variant="outlined" sx={{ p: 3 }}>
        <Box component="form" onSubmit={handleSubmit(onSubmit)} noValidate>
          <OfflineDataNotice
            hasData={Boolean(existingItem) || !isEdit}
            dataUpdatedAt={dataUpdatedAt}
          />

          {isOffline && (
            <Alert severity="warning" sx={{ mb: 2 }}>
              {OFFLINE_MUTATION_MESSAGE} Catalog edits are queued locally and
              will sync once connectivity returns.
            </Alert>
          )}

          {error && isEdit && (
            <Alert severity="warning" sx={{ mb: 2 }}>
              Catalog sync is temporarily unavailable. Showing the latest cached
              draft details when possible.
            </Alert>
          )}

          {errors.root && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {errors.root.message}
            </Alert>
          )}

          <Typography variant="subtitle2" fontWeight={600} gutterBottom>
            Basic Information
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            New catalog items are saved as drafts until they are explicitly
            published.
          </Typography>
          <Grid container spacing={2} sx={{ mb: 3 }}>
            <Grid item xs={12} md={6}>
              <TextField
                {...register("name")}
                label="Name"
                fullWidth
                size="small"
                required
                error={Boolean(errors.name)}
                helperText={errors.name?.message}
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                {...register("sku")}
                label="SKU"
                fullWidth
                size="small"
                error={Boolean(errors.sku)}
                helperText={errors.sku?.message}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                {...register("description")}
                label="Description"
                fullWidth
                size="small"
                multiline
                rows={3}
                error={Boolean(errors.description)}
                helperText={errors.description?.message}
              />
            </Grid>
            <Grid item xs={12} md={4}>
              <TextField
                {...register("category")}
                label="Category"
                fullWidth
                size="small"
                required
                error={Boolean(errors.category)}
                helperText={errors.category?.message}
              />
            </Grid>
            <Grid item xs={12} md={4}>
              <TextField
                {...register("brand")}
                label="Brand"
                fullWidth
                size="small"
                required
                error={Boolean(errors.brand)}
                helperText={errors.brand?.message}
              />
            </Grid>
            <Grid item xs={12} md={4}>
              <Controller
                name="condition"
                control={control}
                render={({ field }) => (
                  <FormControl
                    fullWidth
                    size="small"
                    error={Boolean(errors.condition)}
                  >
                    <InputLabel>Condition *</InputLabel>
                    <Select {...field} label="Condition *">
                      <MenuItem value="new">New</MenuItem>
                      <MenuItem value="open_box">Open Box</MenuItem>
                      <MenuItem value="used">Used</MenuItem>
                    </Select>
                    {errors.condition && (
                      <FormHelperText>
                        {errors.condition.message}
                      </FormHelperText>
                    )}
                  </FormControl>
                )}
              />
            </Grid>
          </Grid>

          <Divider sx={{ my: 2 }} />
          <Typography variant="subtitle2" fontWeight={600} gutterBottom>
            Pricing & Billing
          </Typography>
          <Grid container spacing={2} sx={{ mb: 3 }}>
            <Grid item xs={12} md={4}>
              <Controller
                name="billing_model"
                control={control}
                render={({ field }) => (
                  <FormControl
                    fullWidth
                    size="small"
                    error={Boolean(errors.billing_model)}
                  >
                    <InputLabel>Billing Model *</InputLabel>
                    <Select {...field} label="Billing Model *">
                      <MenuItem value="one_time">One-Time Purchase</MenuItem>
                      <MenuItem value="monthly_rental">Monthly Rental</MenuItem>
                    </Select>
                    {errors.billing_model && (
                      <FormHelperText>
                        {errors.billing_model.message}
                      </FormHelperText>
                    )}
                  </FormControl>
                )}
              />
            </Grid>
            <Grid item xs={12} md={4}>
              <TextField
                {...register("unit_price", { valueAsNumber: true })}
                label="Price ($)"
                fullWidth
                size="small"
                type="number"
                inputProps={{ step: "0.01", min: 0 }}
                error={Boolean(errors.unit_price)}
                helperText={errors.unit_price?.message}
              />
            </Grid>
            <Grid item xs={12} md={4}>
              <TextField
                {...register("refundable_deposit", { valueAsNumber: true })}
                label="Refundable Deposit ($)"
                fullWidth
                size="small"
                type="number"
                inputProps={{ step: "0.01", min: 0 }}
                error={Boolean(errors.refundable_deposit)}
                helperText={errors.refundable_deposit?.message}
              />
            </Grid>
          </Grid>

          <Divider sx={{ my: 2 }} />
          <Typography variant="subtitle2" fontWeight={600} gutterBottom>
            Inventory
          </Typography>
          <Grid container spacing={2} sx={{ mb: 3 }}>
            <Grid item xs={12} md={6}>
              <TextField
                {...register("quantity", { valueAsNumber: true })}
                label="Quantity"
                fullWidth
                size="small"
                type="number"
                inputProps={{ min: 0, step: 1 }}
                error={Boolean(errors.quantity)}
                helperText={errors.quantity?.message}
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                {...register("location_id")}
                label="Location ID"
                fullWidth
                size="small"
                placeholder="Optional location UUID"
                error={Boolean(errors.location_id)}
                helperText={errors.location_id?.message}
              />
            </Grid>
          </Grid>

          <Divider sx={{ my: 2 }} />
          <Typography variant="subtitle2" fontWeight={600} gutterBottom>
            Availability Windows
          </Typography>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mb: 3 }}>
            {availabilityWindows.fields.length === 0 && (
              <Typography variant="body2" color="text.secondary">
                Add one or more availability windows if the item is only
                available during specific hours.
              </Typography>
            )}
            {availabilityWindows.fields.map((field, index) => (
              <Paper key={field.id} variant="outlined" sx={{ p: 2 }}>
                <Grid container spacing={2} alignItems="center">
                  <Grid item xs={12} md={5}>
                    <TextField
                      {...register(
                        `availability_windows.${index}.start_time` as const,
                      )}
                      label="Start"
                      type="datetime-local"
                      fullWidth
                      size="small"
                      InputLabelProps={{ shrink: true }}
                      error={Boolean(
                        errors.availability_windows?.[index]?.start_time,
                      )}
                      helperText={
                        errors.availability_windows?.[index]?.start_time
                          ?.message
                      }
                    />
                  </Grid>
                  <Grid item xs={12} md={5}>
                    <TextField
                      {...register(
                        `availability_windows.${index}.end_time` as const,
                      )}
                      label="End"
                      type="datetime-local"
                      fullWidth
                      size="small"
                      InputLabelProps={{ shrink: true }}
                      error={Boolean(
                        errors.availability_windows?.[index]?.end_time,
                      )}
                      helperText={
                        errors.availability_windows?.[index]?.end_time?.message
                      }
                    />
                  </Grid>
                  <Grid item xs={12} md={2}>
                    <Box
                      sx={{
                        display: "flex",
                        justifyContent: { xs: "flex-end", md: "center" },
                      }}
                    >
                      <IconButton
                        aria-label={`Remove availability window ${index + 1}`}
                        onClick={() => availabilityWindows.remove(index)}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </Box>
                  </Grid>
                </Grid>
              </Paper>
            ))}
            <Button
              variant="outlined"
              startIcon={<AddIcon />}
              onClick={() =>
                availabilityWindows.append({ ...EMPTY_CATALOG_WINDOW })
              }
              sx={{ alignSelf: "flex-start" }}
            >
              Add Availability Window
            </Button>
          </Box>

          <Divider sx={{ my: 2 }} />
          <Typography variant="subtitle2" fontWeight={600} gutterBottom>
            Blackout Windows
          </Typography>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mb: 3 }}>
            {blackoutWindows.fields.length === 0 && (
              <Typography variant="body2" color="text.secondary">
                Use blackout windows to block dates or times when this item
                should not be reservable.
              </Typography>
            )}
            {blackoutWindows.fields.map((field, index) => (
              <Paper key={field.id} variant="outlined" sx={{ p: 2 }}>
                <Grid container spacing={2} alignItems="center">
                  <Grid item xs={12} md={5}>
                    <TextField
                      {...register(
                        `blackout_windows.${index}.start_time` as const,
                      )}
                      label="Start"
                      type="datetime-local"
                      fullWidth
                      size="small"
                      InputLabelProps={{ shrink: true }}
                      error={Boolean(
                        errors.blackout_windows?.[index]?.start_time,
                      )}
                      helperText={
                        errors.blackout_windows?.[index]?.start_time?.message
                      }
                    />
                  </Grid>
                  <Grid item xs={12} md={5}>
                    <TextField
                      {...register(
                        `blackout_windows.${index}.end_time` as const,
                      )}
                      label="End"
                      type="datetime-local"
                      fullWidth
                      size="small"
                      InputLabelProps={{ shrink: true }}
                      error={Boolean(
                        errors.blackout_windows?.[index]?.end_time,
                      )}
                      helperText={
                        errors.blackout_windows?.[index]?.end_time?.message
                      }
                    />
                  </Grid>
                  <Grid item xs={12} md={2}>
                    <Box
                      sx={{
                        display: "flex",
                        justifyContent: { xs: "flex-end", md: "center" },
                      }}
                    >
                      <IconButton
                        aria-label={`Remove blackout window ${index + 1}`}
                        onClick={() => blackoutWindows.remove(index)}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </Box>
                  </Grid>
                </Grid>
              </Paper>
            ))}
            <Button
              variant="outlined"
              startIcon={<AddIcon />}
              onClick={() =>
                blackoutWindows.append({ ...EMPTY_CATALOG_WINDOW })
              }
              sx={{ alignSelf: "flex-start" }}
            >
              Add Blackout Window
            </Button>
          </Box>

          <Box sx={{ display: "flex", gap: 2, justifyContent: "flex-end" }}>
            <Button
              variant="outlined"
              onClick={() =>
                navigate(isEdit && id ? `/catalog/${id}` : "/catalog")
              }
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="contained"
              disabled={isSubmitting}
              startIcon={
                isSubmitting ? (
                  <CircularProgress size={16} color="inherit" />
                ) : undefined
              }
            >
              {isEdit ? "Save Changes" : "Save Draft"}
            </Button>
          </Box>
        </Box>
      </Paper>
    </PageContainer>
  );
}
