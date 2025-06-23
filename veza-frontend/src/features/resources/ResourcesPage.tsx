import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { 
  Search, 
  Filter, 
  Download, 
  Share2, 
  Bookmark, 
  MoreVertical,
  FileText,
  Video,
  Code,
  Image,
  Star,
  Clock,
  TrendingUp,
  Plus
} from "lucide-react";

interface Resource {
  id: number;
  title: string;
  description: string;
  type: string;
  category: string;
  tags: string[];
  author: {
    name: string;
    avatar?: string;
  };
  downloads: number;
  rating: number;
  createdAt: string;
  size: string;
  format: string;
}

export function ResourcesPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedType, setSelectedType] = useState<string | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState("recent");
  const [showUploadDialog, setShowUploadDialog] = useState(false);

  // Données de test
  const resources: Resource[] = [
    {
      id: 1,
      title: "Guide de démarrage",
      description: "Un guide complet pour commencer avec notre plateforme",
      type: "document",
      category: "guides",
      tags: ["guide", "démarrage", "tutoriel"],
      author: {
        name: "John Doe",
        avatar: "https://github.com/shadcn.png"
      },
      downloads: 1234,
      rating: 4.5,
      createdAt: "2024-03-15",
      size: "2.5 MB",
      format: "PDF"
    },
    {
      id: 2,
      title: "Template de projet React",
      description: "Un template prêt à l'emploi pour vos projets React",
      type: "template",
      category: "templates",
      tags: ["template", "react", "boilerplate"],
      author: {
        name: "Alice Smith",
        avatar: "https://github.com/shadcn.png"
      },
      downloads: 856,
      rating: 4.8,
      createdAt: "2024-03-14",
      size: "5.1 MB",
      format: "ZIP"
    },
    {
      id: 3,
      title: "Vidéo tutorielle TypeScript",
      description: "Apprenez les bases de TypeScript en 10 minutes",
      type: "video",
      category: "tutorials",
      tags: ["vidéo", "typescript", "formation"],
      author: {
        name: "Bob Johnson",
        avatar: "https://github.com/shadcn.png"
      },
      downloads: 2345,
      rating: 4.2,
      createdAt: "2024-03-13",
      size: "150 MB",
      format: "MP4"
    }
  ];

  const resourceTypes = [
    { id: "all", label: "Tous", icon: Filter },
    { id: "document", label: "Documents", icon: FileText },
    { id: "template", label: "Templates", icon: Code },
    { id: "video", label: "Vidéos", icon: Video },
    { id: "image", label: "Images", icon: Image }
  ];

  const categories = [
    { id: "all", label: "Toutes les catégories" },
    { id: "guides", label: "Guides" },
    { id: "templates", label: "Templates" },
    { id: "tutorials", label: "Tutoriels" },
    { id: "tools", label: "Outils" }
  ];

  const sortOptions = [
    { id: "recent", label: "Plus récent", icon: Clock },
    { id: "popular", label: "Plus populaire", icon: TrendingUp },
    { id: "rating", label: "Mieux noté", icon: Star }
  ];

  const filteredResources = resources
    .filter(resource => {
      const matchesSearch = resource.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          resource.description.toLowerCase().includes(searchQuery.toLowerCase());
      const matchesType = selectedType === null || selectedType === "all" || resource.type === selectedType;
      const matchesCategory = selectedCategory === null || selectedCategory === "all" || resource.category === selectedCategory;
      return matchesSearch && matchesType && matchesCategory;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case "popular":
          return b.downloads - a.downloads;
        case "rating":
          return b.rating - a.rating;
        default:
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
      }
    });

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Ressources</h1>
        <Dialog open={showUploadDialog} onOpenChange={setShowUploadDialog}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Partager une ressource
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Partager une ressource</DialogTitle>
              <DialogDescription>
                Remplissez les informations ci-dessous pour partager votre ressource.
              </DialogDescription>
            </DialogHeader>
            {/* Formulaire de partage à implémenter */}
          </DialogContent>
        </Dialog>
      </div>

      {/* Filtres et recherche */}
      <div className="flex flex-col md:flex-row gap-4">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Rechercher des ressources..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
        </div>
        <div className="flex gap-2">
          <Select value={selectedType || "all"} onValueChange={setSelectedType}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Type de ressource" />
            </SelectTrigger>
            <SelectContent>
              {resourceTypes.map((type) => (
                <SelectItem key={type.id} value={type.id}>
                  <div className="flex items-center">
                    <type.icon className="h-4 w-4 mr-2" />
                    {type.label}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={selectedCategory || "all"} onValueChange={setSelectedCategory}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Catégorie" />
            </SelectTrigger>
            <SelectContent>
              {categories.map((category) => (
                <SelectItem key={category.id} value={category.id}>
                  {category.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={sortBy} onValueChange={setSortBy}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Trier par" />
            </SelectTrigger>
            <SelectContent>
              {sortOptions.map((option) => (
                <SelectItem key={option.id} value={option.id}>
                  <div className="flex items-center">
                    <option.icon className="h-4 w-4 mr-2" />
                    {option.label}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Grille de ressources */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {filteredResources.map((resource) => (
          <Card key={resource.id} className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  {resource.type === "document" && <FileText className="h-5 w-5 text-blue-500" />}
                  {resource.type === "template" && <Code className="h-5 w-5 text-green-500" />}
                  {resource.type === "video" && <Video className="h-5 w-5 text-red-500" />}
                  {resource.type === "image" && <Image className="h-5 w-5 text-purple-500" />}
                  <CardTitle className="text-lg">{resource.title}</CardTitle>
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreVertical className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem>
                      <Bookmark className="h-4 w-4 mr-2" />
                      Sauvegarder
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <Share2 className="h-4 w-4 mr-2" />
                      Partager
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
              <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                <Avatar className="h-6 w-6">
                  <AvatarImage src={resource.author.avatar} alt={resource.author.name} />
                  <AvatarFallback>{resource.author.name[0]}</AvatarFallback>
                </Avatar>
                <span>{resource.author.name}</span>
                <span>•</span>
                <span>{resource.createdAt}</span>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                {resource.description}
              </p>
              <div className="flex flex-wrap gap-2 mb-4">
                {resource.tags.map((tag) => (
                  <Badge key={tag} variant="secondary">
                    {tag}
                  </Badge>
                ))}
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                  <div className="flex items-center">
                    <Download className="h-4 w-4 mr-1" />
                    {resource.downloads}
                  </div>
                  <div className="flex items-center">
                    <Star className="h-4 w-4 mr-1 text-yellow-500" />
                    {resource.rating}
                  </div>
                  <div>{resource.size}</div>
                </div>
                <Button>
                  <Download className="h-4 w-4 mr-2" />
                  Télécharger
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
} 