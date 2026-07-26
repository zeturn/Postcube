import type { CSSProperties, ReactNode } from 'react';
import { Card as WcCard } from '@zeturn/watercolor-react';

interface CardProps {
  children: ReactNode;
  className?: string;
  variant?: 'default' | 'outlined' | 'minimal' | 'elevated';
  style?: CSSProperties;
  onClick?: () => void;
}

export default function Card({ children, className = '', variant = 'outlined', style, onClick }: CardProps) {
  return (
    <WcCard
      variant={variant}
      interactive={!!onClick}
      onClick={onClick}
      className={className}
      style={style}
    >
      {children}
    </WcCard>
  );
}
