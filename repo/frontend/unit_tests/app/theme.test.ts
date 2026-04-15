import { describe, it, expect } from 'vitest';
import theme from '@/app/theme';

describe('theme', () => {
  it('exports a light MUI theme', () => {
    expect(theme).toBeDefined();
    expect(theme.palette.mode).toBe('light');
  });

  it('configures primary and secondary palettes', () => {
    expect(theme.palette.primary.main).toBe('#1B5E20');
    expect(theme.palette.secondary.main).toBe('#FF6F00');
  });

  it('configures status palettes', () => {
    expect(theme.palette.error.main).toBe('#D32F2F');
    expect(theme.palette.warning.main).toBe('#ED6C02');
    expect(theme.palette.success.main).toBe('#2E7D32');
    expect(theme.palette.info.main).toBe('#0288D1');
  });

  it('configures typography with Inter font family', () => {
    expect(theme.typography.fontFamily).toContain('Inter');
    expect(theme.typography.button.textTransform).toBe('none');
  });

  it('sets borderRadius to 8', () => {
    expect(theme.shape.borderRadius).toBe(8);
  });

  it('configures component defaults', () => {
    expect(theme.components?.MuiTextField?.defaultProps).toMatchObject({
      variant: 'outlined',
      size: 'small',
    });
  });
});
