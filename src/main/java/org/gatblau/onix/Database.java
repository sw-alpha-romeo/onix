package org.gatblau.onix;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;

@Service
public class Database {
    private Connection conn;
    private PreparedStatement stmt;
    private String resultKey;

    @Autowired
    private DataSourceFactory ds;

    public Database() {
    }

    public Connection createConnection() throws SQLException {
        conn = ds.instance().getConnection();
        return conn;
    }

    public void prepare(String sql) throws SQLException {
        if (conn == null) {
            createConnection();
        }
        stmt = conn.prepareStatement(sql);
    }

    public void setString(int parameterIndex, String value) throws SQLException {
        stmt.setString(parameterIndex, value);
    }

    public void setInt(int parameterIndex, Integer value) throws SQLException {
        stmt.setInt(parameterIndex, value);
    }

    public void setObject(int parameterIndex, Object value) throws SQLException {
        stmt.setObject(parameterIndex, value);
    }

    public void setObjectRange(int fromIndex, int toIndex, Object value) throws SQLException {
        for (int i = fromIndex; i < toIndex + 1; i++) {
            setObject(i, null);
        }
    }

    public String executeQueryAndRetrieveStatus(String query_name) throws SQLException {
        String result = null;
        ResultSet set = stmt.executeQuery();
        if (set.next()) {
            return set.getString(query_name);
        }
        throw new RuntimeException(String.format("Failed to execute query '%s'", query_name));
    }

    public ResultSet executeQuery() throws SQLException {
        String result = null;
        return stmt.executeQuery();
    }

    public ResultSet executeQuerySingleRow() throws SQLException {
        String result = null;
        ResultSet set = stmt.executeQuery();
        if (set.next()) {
            return set;
        }
        throw new RuntimeException("No results found.");
    }

    public void close() {
        try {
            stmt.close();
            conn.close();
            stmt = null;
            conn = null;
        }
        catch (Exception ex) {
            System.out.println("WARNING: failed to close database connection.");
            ex.printStackTrace();
        }
    }

    public static final String GET_ITEM_SQL = "SELECT * FROM item(?::character varying)";

    public static final String SET_ITEM_SQL =
        "SELECT set_item(" +
            "?::character varying," +
            "?::character varying," +
            "?::text," +
            "?::jsonb," +
            "?::text[]," +
            "?::hstore," +
            "?::smallint," +
            "?::character varying," +
            "?::bigint," +
            "?::character varying)";

    public static final String FIND_LINKS_SQL =
        "SELECT * FROM find_links(" +
            "?::character varying," +
            "?::character varying," +
            "?::text[]," +
            "?::hstore," +
            "?::character varying," +
            "?::timestamp with time zone," +
            "?::timestamp with time zone," +
            "?::timestamp with time zone," +
            "?::timestamp with time zone)";


}
