import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ProductCard } from '../ProductCard';

const mockProduct = {
  id: 1,
  name: 'Test Product',
  description: 'Test Description',
  price: 99.99,
  image: 'https://example.com/image.jpg',
  category: 'Test Category',
  stock: 10,
};

describe('ProductCard', () => {
  it('renders product information correctly', () => {
    render(<ProductCard product={mockProduct} />);

    expect(screen.getByText('Test Product')).toBeInTheDocument();
    expect(screen.getByText('Test Description')).toBeInTheDocument();
    expect(screen.getByText('99,99 €')).toBeInTheDocument();
    expect(screen.getByText('Test Category')).toBeInTheDocument();
    expect(screen.getByText('10 en stock')).toBeInTheDocument();
  });

  it('calls onEdit when edit button is clicked', () => {
    const onEdit = vi.fn();
    render(<ProductCard product={mockProduct} onEdit={onEdit} />);

    fireEvent.click(screen.getByText('Modifier'));
    expect(onEdit).toHaveBeenCalledWith(mockProduct.id);
  });

  it('calls onDelete when delete button is clicked', () => {
    const onDelete = vi.fn();
    render(<ProductCard product={mockProduct} onDelete={onDelete} />);

    fireEvent.click(screen.getByText('Supprimer'));
    expect(onDelete).toHaveBeenCalledWith(mockProduct.id);
  });

  it('shows out of stock badge when stock is 0', () => {
    const outOfStockProduct = { ...mockProduct, stock: 0 };
    render(<ProductCard product={outOfStockProduct} />);

    expect(screen.getByText('Rupture de stock')).toBeInTheDocument();
  });

  it('handles image loading error', () => {
    render(<ProductCard product={mockProduct} />);
    const image = screen.getByAltText('Test Product');
    
    fireEvent.error(image);
    expect(image).toHaveAttribute('src', expect.stringContaining('placeholder'));
  });
}); 