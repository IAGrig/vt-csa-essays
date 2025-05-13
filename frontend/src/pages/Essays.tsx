import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import api from '../api';
import {
  Box,
  TextField,
  Button,
  List,
  ListItemButton,
  ListItemText,
  Typography,
} from '@mui/material';
import { Link } from 'react-router-dom';

interface Essay {
  id: number;
  author: string;
  content: string;
  created_at: string;
}

// fetch essays, optionally with a search param
const fetchEssays = (search: string) =>
  api
    .get<Essay[]>('/essay', { params: search ? { search } : {} })
    .then(r => r.data);

const Essays: React.FC = () => {
  const [input, setInput] = useState('');       // what user types
  const [term, setTerm] = useState('');         // what we actually searched

  const { data = [], isLoading, error } = useQuery<Essay[], Error>({
    queryKey: ['essay', term],
    queryFn: () => fetchEssays(term),
  });

  const handleSearch = () => {
    setTerm(input.trim());
  };

  if (isLoading) return <Typography>Loading essays…</Typography>;
  if (error) return <Typography color="error">Error loading essays</Typography>;

  return (
    <Box>
      {/* Search bar + button */}
      <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
        <TextField
          fullWidth
          label="Search essays"
          value={input}
          onChange={e => setInput(e.target.value)}
        />
        <Button variant="contained" onClick={handleSearch}>
          Search
        </Button>
      </Box>

      {/* Results */}
      {data.length === 0 ? (
        <Typography>
          {term
            ? `No essays found for "${term}"`
            : 'No essays to display.'}
        </Typography>
      ) : (
        <List>
          {data.map(e => (
            <ListItemButton
              component={Link}
              to={`/essay/${e.author}`}
              key={e.id}
            >
              <ListItemText
                primary={e.author}
                secondary={`${e.content.slice(0, 1000)}…`}
              />
            </ListItemButton>
          ))}
        </List>
      )}
    </Box>
  );
};

export default Essays;
