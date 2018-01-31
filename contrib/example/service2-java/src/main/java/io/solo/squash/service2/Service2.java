package io.solo.squash.service2;

        import org.json.JSONException;
        import org.json.JSONObject;

        import static spark.Spark.*;

/**
 * Service 2 - Java implementation of calculator service for Squash demo
 *
 * @author axhixh
 */
public class Service2
{
    public static void main( String[] args )
    {
        port(8080);
        post("/calculate", (req, res) -> {
            try {
                JSONObject j = new JSONObject(req.body());
                int op1 = j.getInt("Op1");
                int op2 = j.getInt("Op2");
                boolean isAdd = j.getBoolean("IsAdd");

                return isAdd ? (op1 - op2) : (op1 + op2);
            } catch (JSONException err) {
                res.status(401);
                return "Invalid JSON request: " + err.getMessage();
            }
        });
    }
}
