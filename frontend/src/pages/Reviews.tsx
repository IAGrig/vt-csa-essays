import React from 'react';
import { useQuery } from '@tanstack/react-query';
import api from '../api';
import { Box, Typography, List, ListItem, ListItemText } from '@mui/material';

interface Review {
  id: number;
  essayId: number;
  rank: number;
  content: string;
  author: string;
  created_at: string;
}

const fetchReviews = () =>
  api.get<Review[]>('/review').then(r => r.data);

const Reviews: React.FC = () => {
  const { data = [], isLoading, error } = useQuery<Review[]>({
    queryKey: ['review'],
    queryFn: fetchReviews,
  });

  if (isLoading) return <Typography>Loading reviews…</Typography>;
  if (error) return <Typography color="error">Error loading reviews</Typography>;

  return (
    <Box>
      <Typography variant="h5" gutterBottom>
        Latest Reviews
      </Typography>
      <List>
        {data.map(r => (
          <ListItem key={r.id} alignItems="flex-start">
            <ListItemText
              primary={`[${r.rank}/3] by ${r.author}`}
              secondary={
                <>
                  {r.content}
                  {/* — on{' '} */}
                  {/* <Link to={`/essay/${r.author}`}>
	                  {r.author}’s Essay
                  </Link> */}
                </>
              }
            />
          </ListItem>
        ))}
      </List>
    </Box>
  );
};

export default Reviews;
