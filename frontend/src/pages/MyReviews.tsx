import React from 'react';
import { useQuery } from '@tanstack/react-query';
import api from '../api';
import { List, ListItem, ListItemText, Typography } from '@mui/material';

interface Review {
  id: number;
  essayId: number;
  rank: number;
  content: string;
  created_at: string;
}

const fetchMyReviews = (author: string) =>
  api.get<Review[]>(`/reviews?author=${author}`).then(r => r.data);

const MyReviews: React.FC = () => {
  const author = localStorage.getItem('username')!;
  const { data = [], isLoading, error } = useQuery<Review[]>({
    queryKey: ['myReviews', author],
    queryFn: () => fetchMyReviews(author),
  });

  if (isLoading) return <Typography>Loading your reviews…</Typography>;
  if (error) return <Typography color="error">Error loading reviews</Typography>;

  return (
    <>
      <Typography variant="h5">My Reviews</Typography>
      {data.length === 0 && <Typography>You haven’t written any reviews yet.</Typography>}
      <List>
        {data.map(r => (
          <ListItem key={r.id}>
            <ListItemText
              primary={`[${r.rank}/3] on essay #${r.essayId}`}
              secondary={r.content}
            />
          </ListItem>
        ))}
      </List>
    </>
  );
};

export default MyReviews;
