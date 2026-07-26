import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Button as WcButton, type ButtonProps as WcButtonProps } from '@zeturn/watercolor-react';

type ButtonVariant = 'brand' | 'dark' | 'outline' | 'ghost' | 'danger';
type ButtonSize = 'sm' | 'md' | 'lg';

interface ButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'color'> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  icon?: ReactNode;
  fullWidth?: boolean;
  children?: ReactNode;
}

// Map Postcube's button vocabulary onto watercolor's variant + buttonStyle contract.
const styleMap: Record<ButtonVariant, { variant: WcButtonProps['variant']; buttonStyle: 'default' | 'outlined' | 'filled' }> = {
  brand: { variant: 'orange', buttonStyle: 'filled' },
  dark: { variant: 'secondary', buttonStyle: 'filled' },
  outline: { variant: 'secondary', buttonStyle: 'outlined' },
  ghost: { variant: 'text', buttonStyle: 'default' },
  danger: { variant: 'error', buttonStyle: 'outlined' },
};

export default function Button({
  variant = 'brand',
  size = 'md',
  loading = false,
  icon,
  fullWidth = false,
  children,
  className = '',
  ...props
}: ButtonProps) {
  const mapped = styleMap[variant];
  return (
    <WcButton
      variant={mapped.variant}
      buttonStyle={mapped.buttonStyle}
      size={size}
      loading={loading}
      fullWidth={fullWidth}
      startIcon={icon}
      className={className}
      {...props}
    >
      {children}
    </WcButton>
  );
}
