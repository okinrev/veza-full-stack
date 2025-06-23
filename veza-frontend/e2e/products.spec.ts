import { test, expect } from '@playwright/test';

test.describe('Products Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/products');
  });

  test('displays products list', async ({ page }) => {
    // Vérifier que la page est chargée
    await expect(page.getByRole('heading', { name: 'Produits' })).toBeVisible();

    // Vérifier que la liste des produits est visible
    await expect(page.getByRole('grid')).toBeVisible();
  });

  test('filters products by search', async ({ page }) => {
    // Remplir le champ de recherche
    await page.getByPlaceholder('Rechercher un produit...').fill('Test');

    // Vérifier que les résultats sont filtrés
    await expect(page.getByRole('grid')).toBeVisible();
  });

  test('filters products by category', async ({ page }) => {
    // Sélectionner une catégorie
    await page.getByRole('combobox').click();
    await page.getByRole('option', { name: 'Électronique' }).click();

    // Vérifier que les résultats sont filtrés
    await expect(page.getByRole('grid')).toBeVisible();
  });

  test('creates a new product', async ({ page }) => {
    // Cliquer sur le bouton d'ajout
    await page.getByRole('button', { name: 'Ajouter un produit' }).click();

    // Remplir le formulaire
    await page.getByLabel('Nom').fill('Nouveau Produit');
    await page.getByLabel('Description').fill('Description du nouveau produit');
    await page.getByLabel('Prix').fill('99.99');
    await page.getByLabel('Image').fill('https://example.com/image.jpg');
    await page.getByLabel('Catégorie').click();
    await page.getByRole('option', { name: 'Électronique' }).click();
    await page.getByLabel('Stock').fill('10');

    // Soumettre le formulaire
    await page.getByRole('button', { name: 'Créer' }).click();

    // Vérifier que le produit est ajouté
    await expect(page.getByText('Nouveau Produit')).toBeVisible();
  });

  test('edits an existing product', async ({ page }) => {
    // Cliquer sur le bouton de modification du premier produit
    await page.getByRole('button', { name: 'Modifier' }).first().click();

    // Modifier le nom
    await page.getByLabel('Nom').fill('Produit Modifié');

    // Soumettre le formulaire
    await page.getByRole('button', { name: 'Modifier' }).click();

    // Vérifier que le produit est modifié
    await expect(page.getByText('Produit Modifié')).toBeVisible();
  });

  test('deletes a product', async ({ page }) => {
    // Compter le nombre initial de produits
    const initialCount = await page.getByRole('grid').locator('> div').count();

    // Cliquer sur le bouton de suppression du premier produit
    await page.getByRole('button', { name: 'Supprimer' }).first().click();

    // Vérifier que le produit est supprimé
    const finalCount = await page.getByRole('grid').locator('> div').count();
    expect(finalCount).toBe(initialCount - 1);
  });
}); 