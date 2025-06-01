// Represents a character entity returned from the API
export interface Character {
  id: number;
  name: string;
  name_japanese: string;
  main_image: string;
  description: string;
  species: string;
  gender: string;
  age: number;
  height: string;
  status: string;
  affiliation: string;
  occupation: string;
  birth_date: string;
  birth_place: string;
  relatives: string;
  first_appearance: number;
}

