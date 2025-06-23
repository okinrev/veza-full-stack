import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { TracksPage } from '../TracksPage';
import { audioService } from '../../services/audioService';
import '@testing-library/jest-dom';

// Mock des services
vi.mock('../../services/audioService');

// Mock des composants enfants
vi.mock('../../components/AudioPlayer', () => ({
  AudioPlayer: ({ title, artist }: { title: string; artist: string }) => (
    <div data-testid="audio-player">
      <div>{title}</div>
      <div>{artist}</div>
    </div>
  ),
}));

vi.mock('../../components/WaveformDisplay', () => ({
  WaveformDisplay: () => <div data-testid="waveform-display" />,
}));

vi.mock('../../components/TrackList', () => ({
  TrackList: ({ tracks, onPlay }: { tracks: any[]; onPlay: (track: any) => void }) => (
    <div data-testid="track-list">
      {tracks.map((track) => (
        <button
          key={track.id}
          data-testid={`track-${track.id}`}
          onClick={() => onPlay(track)}
        >
          {track.title}
        </button>
      ))}
    </div>
  ),
}));

vi.mock('../../components/TrackUpload', () => ({
  TrackUpload: ({ onUploadComplete }: { onUploadComplete: () => void }) => (
    <div data-testid="track-upload">
      <button onClick={onUploadComplete}>Upload Complete</button>
    </div>
  ),
}));

vi.mock('../../components/PlaylistManager', () => ({
  PlaylistManager: () => <div data-testid="playlist-manager" />,
}));

describe('TracksPage', () => {
  const mockTracks = [
    {
      id: 1,
      title: 'Test Track 1',
      artist: 'Artist 1',
      album: 'Album 1',
      filename: 'track1.mp3',
      duration_seconds: 180,
      tags: ['tag1', 'tag2'],
      is_public: true,
      uploader_id: 1,
      uploader_username: 'user1',
      play_count: 0,
      created_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 2,
      title: 'Test Track 2',
      artist: 'Artist 2',
      album: 'Album 2',
      filename: 'track2.mp3',
      duration_seconds: 240,
      tags: ['tag3', 'tag4'],
      is_public: true,
      uploader_id: 2,
      uploader_username: 'user2',
      play_count: 0,
      created_at: '2024-01-02T00:00:00Z',
    },
  ];

  beforeEach(() => {
    // Reset des mocks
    vi.clearAllMocks();
    
    // Mock de getTracks
    vi.mocked(audioService.getTracks).mockResolvedValue({
      tracks: mockTracks,
      total: 2,
      page: 1,
      limit: 10,
    });

    // Mock de getStreamUrl
    vi.mocked(audioService.getStreamUrl).mockResolvedValue({
      url: 'http://example.com/stream',
    });

    // Mock de incrementPlayCount
    vi.mocked(audioService.incrementPlayCount).mockResolvedValue(undefined);
  });

  it('devrait charger et afficher la liste des pistes', async () => {
    render(<TracksPage />);

    // Vérifier l'état de chargement initial
    expect(screen.getByText('Pistes Audio')).toBeInTheDocument();

    // Attendre le chargement des pistes
    await waitFor(() => {
      expect(screen.getByTestId('track-list')).toBeInTheDocument();
    });

    // Vérifier que les pistes sont affichées
    expect(screen.getByTestId('track-1')).toHaveTextContent('Test Track 1');
    expect(screen.getByTestId('track-2')).toHaveTextContent('Test Track 2');
  });

  it('devrait gérer les erreurs de chargement', async () => {
    // Simuler une erreur
    vi.mocked(audioService.getTracks).mockRejectedValue(new Error('Erreur de chargement'));

    render(<TracksPage />);

    // Vérifier le message d'erreur
    await waitFor(() => {
      expect(screen.getByText('Erreur lors du chargement des pistes')).toBeInTheDocument();
    });
  });

  it('devrait filtrer les pistes lors de la recherche', async () => {
    render(<TracksPage />);

    // Attendre le chargement des pistes
    await waitFor(() => {
      expect(screen.getByTestId('track-list')).toBeInTheDocument();
    });

    // Rechercher une piste
    const searchInput = screen.getByPlaceholderText('Rechercher des pistes...');
    await userEvent.type(searchInput, 'Track 1');

    // Vérifier que seule la piste correspondante est affichée
    expect(screen.getByTestId('track-1')).toBeInTheDocument();
    expect(screen.queryByTestId('track-2')).not.toBeInTheDocument();
  });

  it('devrait jouer une piste lors du clic', async () => {
    render(<TracksPage />);

    // Attendre le chargement des pistes
    await waitFor(() => {
      expect(screen.getByTestId('track-list')).toBeInTheDocument();
    });

    // Cliquer sur une piste
    fireEvent.click(screen.getByTestId('track-1'));

    // Vérifier que le lecteur audio est affiché
    await waitFor(() => {
      expect(screen.getByTestId('audio-player')).toBeInTheDocument();
    });

    // Vérifier que les informations de la piste sont affichées
    expect(screen.getByText('Test Track 1')).toBeInTheDocument();
    expect(screen.getByText('Artist 1')).toBeInTheDocument();
  });

  it('devrait afficher le formulaire de téléchargement', async () => {
    render(<TracksPage />);

    // Cliquer sur le bouton de téléchargement
    fireEvent.click(screen.getByText('Télécharger'));

    // Vérifier que le formulaire est affiché
    expect(screen.getByTestId('track-upload')).toBeInTheDocument();
  });

  it('devrait afficher le gestionnaire de playlists', async () => {
    render(<TracksPage />);

    // Cliquer sur le bouton des playlists
    fireEvent.click(screen.getByText('Playlists'));

    // Vérifier que le gestionnaire est affiché
    expect(screen.getByTestId('playlist-manager')).toBeInTheDocument();
  });

  it('devrait recharger les pistes après un téléchargement', async () => {
    render(<TracksPage />);

    // Afficher le formulaire de téléchargement
    fireEvent.click(screen.getByText('Télécharger'));

    // Simuler la fin du téléchargement
    fireEvent.click(screen.getByText('Upload Complete'));

    // Vérifier que getTracks est appelé à nouveau
    await waitFor(() => {
      expect(audioService.getTracks).toHaveBeenCalledTimes(2);
    });
  });
}); 