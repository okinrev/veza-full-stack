import { useState, useCallback } from 'react';
import { Upload, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { audioService } from '../services/audioService';
import { cn } from '@/shared/utils/helpers';

interface TrackUploadProps {
  onUploadComplete?: () => void;
  className?: string;
}

export const TrackUpload = ({ onUploadComplete, className }: TrackUploadProps) => {
  const [file, setFile] = useState<File | null>(null);
  const [title, setTitle] = useState('');
  const [artist, setArtist] = useState('');
  const [album, setAlbum] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState('');
  const [isPublic, setIsPublic] = useState(true);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0];
    if (!selectedFile) return;

    if (!selectedFile.type.startsWith('audio/')) {
      setError('Le fichier doit être un fichier audio');
      return;
    }

    setFile(selectedFile);
    setError(null);
  };

  const handleTagInputKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && tagInput.trim()) {
      e.preventDefault();
      if (!tags.includes(tagInput.trim())) {
        setTags([...tags, tagInput.trim()]);
      }
      setTagInput('');
    }
  };

  const removeTag = (tagToRemove: string) => {
    setTags(tags.filter(tag => tag !== tagToRemove));
  };

  const handleUpload = async () => {
    if (!file) {
      setError('Veuillez sélectionner un fichier audio');
      return;
    }

    if (!title.trim()) {
      setError('Le titre est requis');
      return;
    }

    if (!artist.trim()) {
      setError('L\'artiste est requis');
      return;
    }

    setIsUploading(true);
    setError(null);

    try {
      await audioService.uploadTrack(
        file,
        {
          title: title.trim(),
          artist: artist.trim(),
          album: album.trim() || undefined,
          tags,
          is_public: isPublic,
        },
        (progress) => setUploadProgress(progress)
      );

      // Réinitialiser le formulaire
      setFile(null);
      setTitle('');
      setArtist('');
      setAlbum('');
      setTags([]);
      setUploadProgress(0);
      onUploadComplete?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Une erreur est survenue lors du téléchargement');
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <Card className={cn('p-6', className)}>
      <h2 className="text-xl font-semibold mb-6">Télécharger une piste audio</h2>

      <div className="space-y-6">
        {/* Sélection du fichier */}
        <div>
          <label className="block text-sm font-medium mb-2">
            Fichier audio
          </label>
          <div className="flex items-center gap-4">
            <label className="flex-1">
              <input
                type="file"
                accept="audio/*"
                onChange={handleFileChange}
                className="hidden"
                disabled={isUploading}
              />
              <div className="flex items-center justify-center h-32 border-2 border-dashed rounded-lg cursor-pointer hover:border-gray-400 transition-colors">
                {file ? (
                  <div className="text-center">
                    <p className="font-medium">{file.name}</p>
                    <p className="text-sm text-gray-500">
                      {(file.size / 1024 / 1024).toFixed(2)} MB
                    </p>
                  </div>
                ) : (
                  <div className="text-center">
                    <Upload className="h-8 w-8 mx-auto mb-2 text-gray-400" />
                    <p className="text-sm text-gray-500">
                      Cliquez pour sélectionner un fichier audio
                    </p>
                  </div>
                )}
              </div>
            </label>
          </div>
        </div>

        {/* Métadonnées */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">
              Titre *
            </label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Titre de la piste"
              disabled={isUploading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Artiste *
            </label>
            <Input
              value={artist}
              onChange={(e) => setArtist(e.target.value)}
              placeholder="Nom de l'artiste"
              disabled={isUploading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Album
            </label>
            <Input
              value={album}
              onChange={(e) => setAlbum(e.target.value)}
              placeholder="Nom de l'album"
              disabled={isUploading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Tags
            </label>
            <Input
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={handleTagInputKeyDown}
              placeholder="Appuyez sur Entrée pour ajouter un tag"
              disabled={isUploading}
            />
            {tags.length > 0 && (
              <div className="flex flex-wrap gap-2 mt-2">
                {tags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center gap-1 px-2 py-1 bg-gray-100 rounded-full text-sm"
                  >
                    {tag}
                    <button
                      type="button"
                      onClick={() => removeTag(tag)}
                      className="text-gray-500 hover:text-gray-700"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Visibilité */}
        <div>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
              disabled={isUploading}
              className="rounded border-gray-300"
            />
            <span className="text-sm">Rendre public</span>
          </label>
        </div>

        {/* Barre de progression */}
        {isUploading && (
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-blue-600 h-2 rounded-full transition-all duration-300"
              style={{ width: `${uploadProgress}%` }}
            />
          </div>
        )}

        {/* Message d'erreur */}
        {error && (
          <p className="text-sm text-red-600">{error}</p>
        )}

        {/* Bouton de téléchargement */}
        <Button
          onClick={handleUpload}
          disabled={isUploading || !file}
          className="w-full"
        >
          {isUploading ? 'Téléchargement...' : 'Télécharger'}
        </Button>
      </div>
    </Card>
  );
}; 