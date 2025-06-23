import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuthStore } from "@/shared/stores/authStore";
import { 
  User, 
  Mail, 
  Lock, 
  Bell, 
  Activity, 
  Settings, 
  Camera,
  Save,
  LogOut
} from "lucide-react";

export function ProfilePage() {
  const { user, logout } = useAuthStore();
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({
    username: user?.username || "",
    email: user?.email || "",
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    // TODO: Implémenter la mise à jour du profil
    setIsEditing(false);
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Profil</h1>
        <Button variant="outline" onClick={logout}>
          <LogOut className="h-4 w-4 mr-2" />
          Déconnexion
        </Button>
      </div>

      <Tabs defaultValue="profile" className="space-y-6">
        <TabsList>
          <TabsTrigger value="profile">
            <User className="h-4 w-4 mr-2" />
            Profil
          </TabsTrigger>
          <TabsTrigger value="activity">
            <Activity className="h-4 w-4 mr-2" />
            Activité
          </TabsTrigger>
          <TabsTrigger value="settings">
            <Settings className="h-4 w-4 mr-2" />
            Paramètres
          </TabsTrigger>
        </TabsList>

        <TabsContent value="profile" className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Informations personnelles</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-col items-center space-y-4">
                  <div className="relative">
                    <Avatar className="h-24 w-24">
                      <AvatarImage src={user?.avatar} alt={user?.username} />
                      <AvatarFallback>{user?.username?.[0].toUpperCase()}</AvatarFallback>
                    </Avatar>
                    <Button
                      size="icon"
                      variant="secondary"
                      className="absolute bottom-0 right-0 rounded-full"
                    >
                      <Camera className="h-4 w-4" />
                    </Button>
                  </div>
                  <div className="text-center">
                    <h2 className="text-xl font-semibold">{user?.username}</h2>
                    <p className="text-muted-foreground">{user?.email}</p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Statistiques</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">Cours suivis</p>
                    <p className="text-2xl font-bold">12</p>
                  </div>
                  <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">Certificats</p>
                    <p className="text-2xl font-bold">3</p>
                  </div>
                  <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">Heures d'apprentissage</p>
                    <p className="text-2xl font-bold">48</p>
                  </div>
                  <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">Niveau</p>
                    <p className="text-2xl font-bold">Intermédiaire</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="activity" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Activité récente</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="flex items-center space-x-4 p-4 rounded-lg bg-muted">
                    <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                      <Activity className="h-5 w-5 text-primary" />
                    </div>
                    <div>
                      <p className="font-medium">Cours complété : Introduction à React</p>
                      <p className="text-sm text-muted-foreground">Il y a 2 jours</p>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="settings" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Paramètres du compte</CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-6">
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="username">Nom d'utilisateur</Label>
                    <Input
                      id="username"
                      name="username"
                      value={formData.username}
                      onChange={handleInputChange}
                      disabled={!isEditing}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="email">Email</Label>
                    <Input
                      id="email"
                      name="email"
                      type="email"
                      value={formData.email}
                      onChange={handleInputChange}
                      disabled={!isEditing}
                    />
                  </div>
                </div>

                {isEditing && (
                  <div className="space-y-4">
                    <h3 className="font-medium">Changer le mot de passe</h3>
                    <div className="space-y-2">
                      <Label htmlFor="currentPassword">Mot de passe actuel</Label>
                      <Input
                        id="currentPassword"
                        name="currentPassword"
                        type="password"
                        value={formData.currentPassword}
                        onChange={handleInputChange}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="newPassword">Nouveau mot de passe</Label>
                      <Input
                        id="newPassword"
                        name="newPassword"
                        type="password"
                        value={formData.newPassword}
                        onChange={handleInputChange}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="confirmPassword">Confirmer le mot de passe</Label>
                      <Input
                        id="confirmPassword"
                        name="confirmPassword"
                        type="password"
                        value={formData.confirmPassword}
                        onChange={handleInputChange}
                      />
                    </div>
                  </div>
                )}

                <div className="flex justify-end space-x-2">
                  {isEditing ? (
                    <>
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() => setIsEditing(false)}
                      >
                        Annuler
                      </Button>
                      <Button type="submit">
                        <Save className="h-4 w-4 mr-2" />
                        Enregistrer
                      </Button>
                    </>
                  ) : (
                    <Button type="button" onClick={() => setIsEditing(true)}>
                      Modifier le profil
                    </Button>
                  )}
                </div>
              </form>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Notifications</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>Notifications par email</Label>
                    <p className="text-sm text-muted-foreground">
                      Recevoir des notifications par email
                    </p>
                  </div>
                  <Button variant="outline" size="sm">
                    <Bell className="h-4 w-4 mr-2" />
                    Configurer
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
} 