import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import type { Product } from '../services/productService';

const formSchema = z.object({
  name: z.string().min(1, 'Le nom est requis'),
  description: z.string().min(1, 'La description est requise'),
  price: z.number().min(0, 'Le prix doit être positif'),
  image: z.string().url('L\'URL de l\'image n\'est pas valide'),
  category: z.string().min(1, 'La catégorie est requise'),
  stock: z.number().int().min(0, 'Le stock doit être positif'),
});

type FormData = z.infer<typeof formSchema>;

interface ProductFormProps {
  product?: Product | null;
  onSubmit: (data: FormData) => Promise<void>;
  onCancel: () => void;
}

export function ProductForm({ product, onSubmit, onCancel }: ProductFormProps) {
  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: product?.name ?? '',
      description: product?.description ?? '',
      price: product?.price ?? 0,
      image: product?.image ?? '',
      category: product?.category ?? '',
      stock: product?.stock ?? 0,
    },
  });

  const handleSubmit = async (data: FormData) => {
    await onSubmit(data);
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Nom</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Description</FormLabel>
              <FormControl>
                <Textarea {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="price"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Prix</FormLabel>
              <FormControl>
                <Input
                  type="number"
                  step="0.01"
                  {...field}
                  onChange={(e) => field.onChange(parseFloat(e.target.value))}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="image"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Image</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="category"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Catégorie</FormLabel>
              <Select
                value={field.value}
                onValueChange={field.onChange}
              >
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Sélectionner une catégorie" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="Électronique">Électronique</SelectItem>
                  <SelectItem value="Vêtements">Vêtements</SelectItem>
                  <SelectItem value="Alimentation">Alimentation</SelectItem>
                  <SelectItem value="Maison">Maison</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="stock"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Stock</FormLabel>
              <FormControl>
                <Input
                  type="number"
                  {...field}
                  onChange={(e) => field.onChange(parseInt(e.target.value))}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex justify-end gap-4">
          <Button type="button" variant="outline" onClick={onCancel}>
            Annuler
          </Button>
          <Button type="submit">
            {product ? 'Modifier' : 'Créer'}
          </Button>
        </div>
      </form>
    </Form>
  );
} 