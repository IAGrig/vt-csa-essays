import { AppBar, Toolbar, Typography, Button, Box, Menu, MenuItem, IconButton, Avatar } from '@mui/material';
import { Link, useNavigate } from 'react-router-dom';
import React from 'react';
import { clearTokens, getAccessToken } from '../auth';

const Navbar: React.FC = () => {
  const navigate = useNavigate();
  const token = getAccessToken();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

  const handleMenu = (e: React.MouseEvent<HTMLElement>) => setAnchorEl(e.currentTarget);
  const handleClose = () => setAnchorEl(null);
  const logout = () => { clearTokens(); navigate('/login'); };

  return (
    <AppBar position="static">
      <Toolbar>
        <Typography
          component={Link}
          to="/"
          variant="h6"
          sx={{ flexGrow: 1, textDecoration: 'none', color: 'inherit' }}
        >
          {/* <img src="../assets/react.svg" alt="Logo" style={{ height: 30, marginRight: 8 }} /> */}
          VT CSA ESSAYS
        </Typography>
        <Button component={Link} to="/essays" color="inherit">Essays</Button>
        <Button component={Link} to="/reviews" color="inherit">Reviews</Button>
        {token && <Button component={Link} to="/my-essay" color="inherit">My Essay</Button>}
        {/* {token && <Button component={Link} to="/my-reviews" color="inherit">My Reviews</Button>} */}
        {token ? (
          <Box>
            <IconButton size="large" onClick={handleMenu} color="inherit">
              <Avatar />
            </IconButton>
            <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleClose}>
							<MenuItem disabled>{localStorage.getItem('username')!}</MenuItem>
              <MenuItem onClick={logout}>Logout</MenuItem>
            </Menu>
          </Box>
        ) : (
          <>
            <Button component={Link} to="/login" color="inherit">Login</Button>
            <Button component={Link} to="/register" color="inherit">Register</Button>
          </>
        )}
      </Toolbar>
    </AppBar>
  );
};

export default Navbar;
