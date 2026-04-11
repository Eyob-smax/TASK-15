import { useState, useCallback } from 'react';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Button from '@mui/material/Button';
import ClearIcon from '@mui/icons-material/Clear';

export interface FilterOption {
  value: string;
  label: string;
}

export interface FilterField {
  key: string;
  label: string;
  type: 'text' | 'select' | 'date';
  options?: FilterOption[];
  placeholder?: string;
}

type FilterValues = Record<string, string>;

interface FilterBarProps {
  fields: FilterField[];
  onChange: (values: FilterValues) => void;
  initialValues?: FilterValues;
}

export function FilterBar({ fields, onChange, initialValues = {} }: FilterBarProps) {
  const [values, setValues] = useState<FilterValues>(initialValues);

  const handleChange = useCallback(
    (key: string, value: string) => {
      setValues(prev => {
        const next = { ...prev, [key]: value };
        onChange(next);
        return next;
      });
    },
    [onChange],
  );

  const handleClear = useCallback(() => {
    const cleared: FilterValues = {};
    setValues(cleared);
    onChange(cleared);
  }, [onChange]);

  const hasFilters = Object.values(values).some(v => v !== '' && v !== undefined);

  return (
    <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'center' }}>
      {fields.map(field => {
        if (field.type === 'select') {
          const labelId = `${field.key}-filter-label`;
          const selectId = `${field.key}-filter-select`;
          return (
            <FormControl key={field.key} size="small" sx={{ minWidth: 160 }}>
              <InputLabel id={labelId}>{field.label}</InputLabel>
              <Select
                id={selectId}
                labelId={labelId}
                value={values[field.key] ?? ''}
                label={field.label}
                onChange={e => handleChange(field.key, e.target.value)}
              >
                <MenuItem value="">
                  <em>All</em>
                </MenuItem>
                {field.options?.map(opt => (
                  <MenuItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          );
        }

        return (
          <TextField
            key={field.key}
            size="small"
            label={field.label}
            placeholder={field.placeholder}
            type={field.type === 'date' ? 'date' : 'text'}
            value={values[field.key] ?? ''}
            onChange={e => handleChange(field.key, e.target.value)}
            InputLabelProps={field.type === 'date' ? { shrink: true } : undefined}
            sx={{ minWidth: 160 }}
          />
        );
      })}

      {hasFilters && (
        <Button
          size="small"
          variant="text"
          startIcon={<ClearIcon />}
          onClick={handleClear}
          color="inherit"
        >
          Clear
        </Button>
      )}
    </Box>
  );
}
