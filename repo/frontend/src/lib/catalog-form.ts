import type { Item } from '@/lib/types';
import type { CreateItemFormData } from '@/lib/validation';

export const EMPTY_CATALOG_WINDOW = {
  start_time: '',
  end_time: '',
} as const;

export function getDefaultCatalogFormValues(): CreateItemFormData {
  return {
    name: '',
    description: '',
    category: '',
    brand: '',
    sku: '',
    condition: 'new',
    billing_model: 'one_time',
    unit_price: 0,
    refundable_deposit: 50,
    quantity: 0,
    location_id: '',
    availability_windows: [],
    blackout_windows: [],
  };
}

export function mapItemToCatalogFormValues(item: Item): CreateItemFormData {
  return {
    ...getDefaultCatalogFormValues(),
    name: item.name,
    description: item.description,
    category: item.category,
    brand: item.brand,
    sku: item.sku,
    condition: item.condition,
    billing_model: item.billing_model,
    unit_price: item.unit_price,
    refundable_deposit: item.refundable_deposit,
    quantity: item.quantity,
    location_id: item.location_id ?? '',
    availability_windows: (item.availability_windows ?? []).map((window) => ({
      start_time: toDateTimeLocal(window.start_time),
      end_time: toDateTimeLocal(window.end_time),
    })),
    blackout_windows: (item.blackout_windows ?? []).map((window) => ({
      start_time: toDateTimeLocal(window.start_time),
      end_time: toDateTimeLocal(window.end_time),
    })),
  };
}

export function mapCatalogFormToItemPayload(form: CreateItemFormData) {
  const locationID = form.location_id?.trim();

  return {
    sku: form.sku.trim(),
    name: form.name.trim(),
    description: form.description?.trim() ?? '',
    category: form.category.trim(),
    brand: form.brand.trim(),
    condition: form.condition,
    unit_price: form.unit_price,
    refundable_deposit: form.refundable_deposit,
    billing_model: form.billing_model,
    quantity: form.quantity,
    ...(locationID ? { location_id: locationID } : {}),
    availability_windows: form.availability_windows.map((window) => ({
      start_time: new Date(window.start_time).toISOString(),
      end_time: new Date(window.end_time).toISOString(),
    })),
    blackout_windows: form.blackout_windows.map((window) => ({
      start_time: new Date(window.start_time).toISOString(),
      end_time: new Date(window.end_time).toISOString(),
    })),
  };
}

function toDateTimeLocal(value: string): string {
  if (!value) {
    return '';
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '';
  }

  const pad = (part: number) => String(part).padStart(2, '0');

  return [
    date.getFullYear(),
    '-',
    pad(date.getMonth() + 1),
    '-',
    pad(date.getDate()),
    'T',
    pad(date.getHours()),
    ':',
    pad(date.getMinutes()),
  ].join('');
}
